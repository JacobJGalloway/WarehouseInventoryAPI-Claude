using System.Threading.Channels;
using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Logging.Abstractions;
using Xunit;
using Switchyard.InventoryAPI.Data;
using Switchyard.InventoryAPI.Data.Sync;
using Switchyard.Domain;

namespace Switchyard.InventoryAPI.Tests.Data.Sync
{
    // Uses SQLite in-memory (not the EF Core InMemory provider) because
    // ResyncTableAsync calls ExecuteDeleteAsync, which the InMemory provider
    // does not support (throws InvalidOperationException at runtime) — only
    // relational providers implement the bulk ExecuteDelete/ExecuteUpdate APIs.
    public class InventorySyncWorkerTests : IDisposable
    {
        private readonly SqliteConnection _writeConnection;
        private readonly SqliteConnection _readConnection;
        private readonly InventoryContext _writeCtx;
        private readonly InventoryReadContext _readCtx;
        private readonly ServiceProvider _provider;
        private readonly Channel<SyncJob> _channel;

        public InventorySyncWorkerTests()
        {
            _writeConnection = new SqliteConnection("Filename=:memory:");
            _writeConnection.Open();
            _readConnection = new SqliteConnection("Filename=:memory:");
            _readConnection.Open();

            var writeOptions = new DbContextOptionsBuilder<InventoryContext>()
                .UseSqlite(_writeConnection)
                .Options;
            var readOptions = new DbContextOptionsBuilder<InventoryReadContext>()
                .UseSqlite(_readConnection)
                .Options;
            _writeCtx = new InventoryContext(writeOptions);
            _writeCtx.Database.EnsureCreated();
            _readCtx = new InventoryReadContext(readOptions);
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

        private InventorySyncWorker CreateWorker(ILogger<InventorySyncWorker>? logger = null) =>
            new(_channel.Reader, _provider.GetRequiredService<IServiceScopeFactory>(), logger ?? NullLogger<InventorySyncWorker>.Instance);

        /// <summary>Queues jobs, closes the channel, starts the worker, and awaits its natural completion.</summary>
        private static async Task RunToCompletionAsync(InventorySyncWorker worker, ChannelWriter<SyncJob> writer, params SyncJob[] jobs)
        {
            foreach (var job in jobs) await writer.WriteAsync(job);
            writer.Complete();

            await worker.StartAsync(CancellationToken.None);
            await worker.ExecuteTask!;
            await worker.StopAsync(CancellationToken.None);
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsClothing_FromWriteToReadContext()
        {
            _writeCtx.Clothing.AddRange(
                new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow },
                new Clothing { PartitionKey = "WH001-CLTH002-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", SKUMarker = "CLTH002", UnloadedDate = DateTime.UtcNow }
            );
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(Clothing) }));

            Assert.Equal(2, await _readCtx.Clothing.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsPPE_FromWriteToReadContext()
        {
            _writeCtx.PPE.Add(new PPE { PartitionKey = "WH001-PPE001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PPE001", UnloadedDate = DateTime.UtcNow });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(PPE) }));

            Assert.Equal(1, await _readCtx.PPE.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsTool_FromWriteToReadContext()
        {
            _writeCtx.Tool.Add(new Tool { PartitionKey = "WH001-TOOL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "TOOL001", UnloadedDate = DateTime.UtcNow });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(Tool) }));

            Assert.Equal(1, await _readCtx.Tool.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ResyncsAllChangedTypes_WhenJobListsMultiple()
        {
            _writeCtx.Clothing.Add(new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });
            _writeCtx.PPE.Add(new PPE { PartitionKey = "WH001-PPE001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PPE001", UnloadedDate = DateTime.UtcNow });
            _writeCtx.Tool.Add(new Tool { PartitionKey = "WH001-TOOL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "TOOL001", UnloadedDate = DateTime.UtcNow });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer,
                new SyncJob(new HashSet<Type> { typeof(Clothing), typeof(PPE), typeof(Tool) }));

            Assert.Equal(1, await _readCtx.Clothing.CountAsync());
            Assert.Equal(1, await _readCtx.PPE.CountAsync());
            Assert.Equal(1, await _readCtx.Tool.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_ClearsStaleReadRows_BeforeApplyingResync()
        {
            // Read side has a row the write side no longer has (e.g. deleted upstream).
            _readCtx.Clothing.Add(new Clothing { PartitionKey = "WH001-STALE-ffffffffffffffffffffffffffffffff", SKUMarker = "STALE", UnloadedDate = DateTime.UtcNow });
            await _readCtx.SaveChangesAsync();

            _writeCtx.Clothing.Add(new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(Clothing) }));

            var readRows = await _readCtx.Clothing.ToListAsync();
            var single = Assert.Single(readRows);
            Assert.Equal("CLTH001", single.SKUMarker);
        }

        [Fact]
        public async Task ExecuteAsync_ProcessesMultipleQueuedJobs_InOrder()
        {
            _writeCtx.Clothing.Add(new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });
            await _writeCtx.SaveChangesAsync();

            var worker = CreateWorker();
            await RunToCompletionAsync(worker, _channel.Writer,
                new SyncJob(new HashSet<Type> { typeof(Clothing) }),
                new SyncJob(new HashSet<Type> { typeof(PPE) }));

            Assert.Equal(1, await _readCtx.Clothing.CountAsync());
            Assert.Equal(0, await _readCtx.PPE.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_UnrecognizedChangedType_IsNoOp_AndDoesNotThrow()
        {
            var worker = CreateWorker();

            // SKUCatalog isn't handled by the sync switch — should be silently skipped, not an error.
            await RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(SKUCatalog) }));

            Assert.Equal(0, await _readCtx.Clothing.CountAsync());
        }

        [Fact]
        public async Task ExecuteAsync_LogsErrorAndDoesNotThrow_WhenSyncFails()
        {
            var logger = new CapturingLogger<InventorySyncWorker>();
            var worker = CreateWorker(logger);

            // Dispose the read context out from under the worker so SaveChangesAsync throws
            // inside SyncAsync — proves the catch block swallows and logs instead of crashing.
            await _readCtx.DisposeAsync();
            _writeCtx.Clothing.Add(new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });
            await _writeCtx.SaveChangesAsync();

            var exception = await Record.ExceptionAsync(() =>
                RunToCompletionAsync(worker, _channel.Writer, new SyncJob(new HashSet<Type> { typeof(Clothing) })));

            Assert.Null(exception);
            var message = Assert.Single(logger.ErrorMessages);
            Assert.Contains("Clothing", message);
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
