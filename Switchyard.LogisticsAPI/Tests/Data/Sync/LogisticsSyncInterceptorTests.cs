using System.Threading.Channels;
using Microsoft.EntityFrameworkCore;
using Xunit;
using Switchyard.LogisticsAPI.Data;
using Switchyard.LogisticsAPI.Data.Sync;
using Switchyard.Domain;

namespace Switchyard.LogisticsAPI.Tests.Data.Sync
{
    public class LogisticsSyncInterceptorTests : IDisposable
    {
        private readonly LogisticsContext _context;
        private readonly Channel<SyncJob> _channel;

        public LogisticsSyncInterceptorTests()
        {
            _channel = Channel.CreateUnbounded<SyncJob>();
            var options = new DbContextOptionsBuilder<LogisticsContext>()
                .UseInMemoryDatabase(Guid.NewGuid().ToString())
                .AddInterceptors(new LogisticsSyncInterceptor(_channel.Writer))
                .Options;
            _context = new LogisticsContext(options);
            _context.Database.EnsureCreated();
        }

        public void Dispose() => _context.Dispose();

        [Fact]
        public async Task SavedChangesAsync_QueuesSyncJob_WhenEntityAdded()
        {
            _context.Warehouses.Add(new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" });

            await _context.SaveChangesAsync();

            Assert.True(_channel.Reader.TryRead(out var job));
            Assert.Contains(typeof(Warehouse), job!.ChangedTypes);
        }

        [Fact]
        public async Task SavedChangesAsync_QueuesSyncJob_WhenEntityModified()
        {
            var warehouse = new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" };
            _context.Warehouses.Add(warehouse);
            await _context.SaveChangesAsync();
            _channel.Reader.TryRead(out _); // drain the add's job

            warehouse.City = "Shelbyville";
            await _context.SaveChangesAsync();

            Assert.True(_channel.Reader.TryRead(out var job));
            Assert.Contains(typeof(Warehouse), job!.ChangedTypes);
        }

        [Fact]
        public async Task SavedChangesAsync_QueuesSyncJob_WhenEntityDeleted()
        {
            var warehouse = new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" };
            _context.Warehouses.Add(warehouse);
            await _context.SaveChangesAsync();
            _channel.Reader.TryRead(out _); // drain the add's job

            _context.Warehouses.Remove(warehouse);
            await _context.SaveChangesAsync();

            Assert.True(_channel.Reader.TryRead(out var job));
            Assert.Contains(typeof(Warehouse), job!.ChangedTypes);
        }

        [Fact]
        public async Task SavedChangesAsync_DoesNotQueueJob_WhenNoChangesTracked()
        {
            await _context.SaveChangesAsync();

            Assert.False(_channel.Reader.TryRead(out _));
        }

        [Fact]
        public async Task SavedChangesAsync_QueuesSingleJob_CoveringAllChangedTypes_WhenMultipleTypesChangeTogether()
        {
            _context.Warehouses.Add(new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" });
            _context.Stores.Add(new Store { PartitionKey = "ST001", StoreId = "ST001", BaseWarehouseId = "WH001", City = "Springfield", State = "IL" });

            await _context.SaveChangesAsync();

            Assert.True(_channel.Reader.TryRead(out var job));
            Assert.Equal(2, job!.ChangedTypes.Count);
            Assert.Contains(typeof(Warehouse), job.ChangedTypes);
            Assert.Contains(typeof(Store), job.ChangedTypes);
            Assert.False(_channel.Reader.TryRead(out _)); // exactly one job for the whole SaveChanges call
        }
    }
}
