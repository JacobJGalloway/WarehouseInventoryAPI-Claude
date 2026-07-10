using System.Threading.Channels;
using Microsoft.EntityFrameworkCore;
using Xunit;
using Switchyard.InventoryAPI.Data;
using Switchyard.InventoryAPI.Data.Sync;
using Switchyard.Domain;

namespace Switchyard.InventoryAPI.Tests.Data.Sync
{
    public class InventorySyncInterceptorTests : IDisposable
    {
        private readonly InventoryContext _context;
        private readonly Channel<SyncJob> _channel;

        public InventorySyncInterceptorTests()
        {
            _channel = Channel.CreateUnbounded<SyncJob>();
            var options = new DbContextOptionsBuilder<InventoryContext>()
                .UseInMemoryDatabase(Guid.NewGuid().ToString())
                .AddInterceptors(new InventorySyncInterceptor(_channel.Writer))
                .Options;
            _context = new InventoryContext(options);
            _context.Database.EnsureCreated();
        }

        public void Dispose() => _context.Dispose();

        [Fact]
        public async Task SavedChangesAsync_QueuesSyncJob_WhenEntityAdded()
        {
            _context.Clothing.Add(new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });

            await _context.SaveChangesAsync();

            Assert.True(_channel.Reader.TryRead(out var job));
            Assert.Contains(typeof(Clothing), job!.ChangedTypes);
        }

        [Fact]
        public async Task SavedChangesAsync_QueuesSyncJob_WhenEntityModified()
        {
            var item = new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow };
            _context.Clothing.Add(item);
            await _context.SaveChangesAsync();
            _channel.Reader.TryRead(out _); // drain the add's job

            item.UnitPrice = 9.99m;
            await _context.SaveChangesAsync();

            Assert.True(_channel.Reader.TryRead(out var job));
            Assert.Contains(typeof(Clothing), job!.ChangedTypes);
        }

        [Fact]
        public async Task SavedChangesAsync_QueuesSyncJob_WhenEntityDeleted()
        {
            var item = new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow };
            _context.Clothing.Add(item);
            await _context.SaveChangesAsync();
            _channel.Reader.TryRead(out _); // drain the add's job

            _context.Clothing.Remove(item);
            await _context.SaveChangesAsync();

            Assert.True(_channel.Reader.TryRead(out var job));
            Assert.Contains(typeof(Clothing), job!.ChangedTypes);
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
            _context.Clothing.Add(new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });
            _context.PPE.Add(new PPE { PartitionKey = "WH001-PPE001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PPE001", UnloadedDate = DateTime.UtcNow });

            await _context.SaveChangesAsync();

            Assert.True(_channel.Reader.TryRead(out var job));
            Assert.Equal(2, job!.ChangedTypes.Count);
            Assert.Contains(typeof(Clothing), job.ChangedTypes);
            Assert.Contains(typeof(PPE), job.ChangedTypes);
            Assert.False(_channel.Reader.TryRead(out _)); // exactly one job for the whole SaveChanges call
        }
    }
}
