using System.Threading.Channels;
using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Logging.Abstractions;
using Xunit;
using Switchyard.LogisticsAPI.Data;
using Switchyard.LogisticsAPI.Data.Sync;
using Switchyard.Domain;

namespace Switchyard.LogisticsAPI.Tests.Data.Sync
{
    // Uses SQLite in-memory (not the EF Core InMemory provider) because
    // ResyncTableAsync calls ExecuteDeleteAsync, which the InMemory provider
    // does not support (throws InvalidOperationException at runtime) — only
    // relational providers implement the bulk ExecuteDelete/ExecuteUpdate APIs.
    public class LogisticsSyncWorkerTests : IDisposable
    {
        private readonly SqliteConnection _writeConnection;
        private readonly SqliteConnection _readConnection;
        private readonly LogisticsContext _writeCtx;
        private readonly LogisticsReadContext _readCtx;
        private readonly ServiceProvider _provider;
        private readonly Channel<SyncJob> _channel;

        public LogisticsSyncWorkerTests()
        {
            _writeConnection = new SqliteConnection("Filename=:memory:");
            _writeConnection.Open();
            _readConnection = new SqliteConnection("Filename=:memory:");
            _readConnection.Open();

            var writeOptions = new DbContextOptionsBuilder<LogisticsContext>()
                .UseSqlite(_writeConnection)
                .Options;
            var readOptions = new DbContextOptionsBuilder<LogisticsReadContext>()
                .UseSqlite(_readConnection)
                .Options;
            _writeCtx = new LogisticsContext(writeOptions);
            _writeCtx.Database.EnsureCreated();
            _readCtx = new LogisticsReadContext(readOptions);
            _readCtx.Database.EnsureCreated();

            var services = new ServiceCollection();
            services.AddSingleton(_writeCtx);
            services.AddSingleton(_readCtx);
            _provider = services.BuildServiceProvider();

            _channel = Channel.CreateUnbounded<SyncJob>();
        }

        public void Dispose()
        {
            _writeCtx.Dispose();
            _readCtx.Dispose();
            _writeConnection.Dispose();
            _readConnection.Dispose();
            _provider.Dispose();
        }

        private LogisticsSyncWorker CreateWorker(ILogger<LogisticsSyncWorker>? logger = null) =>
            new(_channel.Reader, _provider.GetRequiredService<IServiceScopeFactory>(), logger ?? NullLogger<LogisticsSyncWorker>.Instance);

        /// <summary>Queues jobs, closes the channel, starts the worker, and awaits its natural completion.</summary>
        private static async Task RunToCompletionAsync(LogisticsSyncWorker worker, ChannelWriter<SyncJob> writer, params SyncJob[] jobs)
        {
            foreach (var job in jobs) await writer.WriteAsync(job);
            writer.Complete();

            await worker.StartAsync(CancellationToken.None);
            await worker.ExecuteTask!;
            await worker.StopAsync(CancellationToken.None);
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsBillsOfLading_FromWriteToReadContext()
        {
            _writeCtx.BillsOfLading.AddRange(
                new BillOfLading { PartitionKey = "BOL001", TransactionId = "T001", LineEntries = [] },
                new BillOfLading { PartitionKey = "BOL002", TransactionId = "T002", LineEntries = [] }
            );
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(BillOfLading) }));

            Assert.Equal(2, await _readCtx.BillsOfLading.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsLineEntries_FromWriteToReadContext()
        {
            _writeCtx.LineEntries.Add(new LineEntry { PartitionKey = "LE001", TransactionId = "T001", LocationId = "WH001", SKUMarker = "SKU001", Quantity = 5 });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(LineEntry) }));

            Assert.Equal(1, await _readCtx.LineEntries.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsWarehouses_FromWriteToReadContext()
        {
            _writeCtx.Warehouses.Add(new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(Warehouse) }));

            Assert.Equal(1, await _readCtx.Warehouses.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsStores_FromWriteToReadContext()
        {
            _writeCtx.Stores.Add(new Store { PartitionKey = "ST001", StoreId = "ST001", BaseWarehouseId = "WH001", City = "Springfield", State = "IL" });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(Store) }));

            Assert.Equal(1, await _readCtx.Stores.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsAllChangedTypes_WhenJobListsMultiple()
        {
            _writeCtx.BillsOfLading.Add(new BillOfLading { PartitionKey = "BOL001", TransactionId = "T001", LineEntries = [] });
            _writeCtx.LineEntries.Add(new LineEntry { PartitionKey = "LE001", TransactionId = "T001", LocationId = "WH001", SKUMarker = "SKU001", Quantity = 5 });
            _writeCtx.Warehouses.Add(new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" });
            _writeCtx.Stores.Add(new Store { PartitionKey = "ST001", StoreId = "ST001", BaseWarehouseId = "WH001", City = "Springfield", State = "IL" });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer,
                new SyncJob(new HashSet<Type> { typeof(BillOfLading), typeof(LineEntry), typeof(Warehouse), typeof(Store) }));

            Assert.Equal(1, await _readCtx.BillsOfLading.CountAsync());
            Assert.Equal(1, await _readCtx.LineEntries.CountAsync());
            Assert.Equal(1, await _readCtx.Warehouses.CountAsync());
            Assert.Equal(1, await _readCtx.Stores.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ClearsStaleReadRows_BeforeApplyingResync()
        {
            _readCtx.Warehouses.Add(new Warehouse { WarehouseId = "WH999-STALE", City = "Nowhere", State = "ZZ" });
            await _readCtx.SaveChangesAsync();

            _writeCtx.Warehouses.Add(new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(Warehouse) }));

            var readRows = await _readCtx.Warehouses.ToListAsync();
            var single = Assert.Single(readRows);
            Assert.Equal("WH001", single.WarehouseId);
        }

        [Fact]
        public async Task ExecuteAsync_ProcessesMultipleQueuedJobs_InOrder()
        {
            _writeCtx.Warehouses.Add(new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer,
                new SyncJob(new HashSet<Type> { typeof(Warehouse) }),
                new SyncJob(new HashSet<Type> { typeof(Store) }));

            Assert.Equal(1, await _readCtx.Warehouses.CountAsync());
            Assert.Equal(0, await _readCtx.Stores.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_UnrecognizedChangedType_IsNoOp_AndDoesNotThrow()
        {
            var worker = CreateWorker();

            // A type outside the sync switch (e.g. a future domain model) should be silently
            // skipped, not treated as an error.
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(object) }));

            Assert.Equal(0, await _readCtx.Warehouses.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_LogsErrorAndDoesNotThrow_WhenSyncFails()
        {
            var logger = new CapturingLogger<LogisticsSyncWorker>();
            var worker = CreateWorker(logger);

            // Dispose the read context out from under the worker so SaveChangesAsync throws
            // inside SyncAsync — proves the catch block swallows and logs instead of crashing.
            await _readCtx.DisposeAsync();
            _writeCtx.Warehouses.Add(new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" });
            await _writeCtx.SaveChangesAsync();

            var exception = await Record.ExceptionAsync(() =>
                RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(Warehouse) })));

            Assert.Null(exception);
            var message = Assert.Single(logger.ErrorMessages);
            Assert.Contains("Warehouse", message);
        }

        private sealed class CapturingLogger<T> : ILogger<T>
        {
            public List<string> ErrorMessages { get; } = new();

            public IDisposable? BeginScope<TState>(TState state) where TState : notnull => null;
            public bool IsEnabled(LogLevel logLevel) => true;

            public void Log<TState>(LogLevel logLevel, EventId eventId, TState state, Exception? exception, Func<TState, Exception?, string> formatter)
            {
                if (logLevel == LogLevel.Error)
                    ErrorMessages.Add(formatter(state, exception));
            }
        }
    }
}
