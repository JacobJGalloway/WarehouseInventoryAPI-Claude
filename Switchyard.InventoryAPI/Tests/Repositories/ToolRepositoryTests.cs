using Xunit;
using Moq;
using Microsoft.EntityFrameworkCore;
using Switchyard.InventoryAPI.Data;
using Switchyard.InventoryAPI.Data.Repositories;
using Switchyard.Domain;

namespace Switchyard.InventoryAPI.Tests.Repositories
{
    public class ToolRepositoryTests : IDisposable
    {
        private readonly InventoryContext _context;
        private readonly InventoryReadContext _readContext;
        private readonly ToolRepository _repository;

        public ToolRepositoryTests()
        {
            var dbName = Guid.NewGuid().ToString();
            var writeOptions = new DbContextOptionsBuilder<InventoryContext>()
                .UseInMemoryDatabase(dbName)
                .Options;
            var readOptions = new DbContextOptionsBuilder<InventoryReadContext>()
                .UseInMemoryDatabase(dbName)
                .Options;
            _context = new InventoryContext(writeOptions);
            _context.Database.EnsureCreated();
            _readContext = new InventoryReadContext(readOptions);
            _repository = new ToolRepository(_context, _readContext);
        }

        public void Dispose()
        {
            _context.Dispose();
            _readContext.Dispose();
        }

        [Fact]
        public async Task GetAllAsync_ReturnsAllItems()
        {
            _context.Tool.AddRange(
                new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow },
                new Tool { PartitionKey = "WH001-PWTL002-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", SKUMarker = "PWTL002", UnloadedDate = DateTime.UtcNow }
            );
            await _context.SaveChangesAsync();

            var result = await _repository.GetAllAsync();

            Assert.Equal(2, result.Count());
        }

        [Fact]
        public async Task GetBySKUIdAsync_ReturnsMatchingItems_WhenFound()
        {
            _context.Tool.AddRange(
                new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow },
                new Tool { PartitionKey = "WH001-PWTL001-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow },
                new Tool { PartitionKey = "WH001-PWTL002-c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8", SKUMarker = "PWTL002", UnloadedDate = DateTime.UtcNow }
            );
            await _context.SaveChangesAsync();

            var result = await _repository.GetBySKUIdAsync("PWTL001");

            Assert.Equal(2, result.Count);
            Assert.All(result, t => Assert.Equal("PWTL001", t.SKUMarker));
        }

        [Fact]
        public async Task GetBySKUIdAsync_ReturnsEmptyList_WhenNotFound()
        {
            var result = await _repository.GetBySKUIdAsync("PWTL999");

            Assert.Empty(result);
        }

        [Fact]
        public async Task AddAsync_StagesItem_PersistedAfterSave()
        {
            var item = new Tool { SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow };

            await _repository.AddAsync(item);
            await _context.SaveChangesAsync();

            Assert.Equal(1, await _context.Tool.CountAsync());
        }

        [Fact]
        public async Task UpdateBySKUIdAsync_UpdatesByPartitionKey_WhenMatch()
        {
            var original = new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow };
            var other = new Tool { PartitionKey = "WH001-PWTL001-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow };
            _context.Tool.AddRange(original, other);
            await _context.SaveChangesAsync();

            var updated = new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow.AddDays(1) };
            await _repository.UpdateBySKUIdAsync("PWTL001", updated);
            await _context.SaveChangesAsync();

            var result = await _context.Tool.FindAsync("WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6");
            Assert.Equal(updated.UnloadedDate, result!.UnloadedDate);
        }

        [Fact]
        public async Task UpdateBySKUIdAsync_FallsBackToFirst_WhenNoPartitionKeyMatch()
        {
            var item = new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow };
            _context.Tool.Add(item);
            await _context.SaveChangesAsync();

            var updated = new Tool { PartitionKey = "WH001-PWTL001-ffffffffffffffffffffffffffffffff", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow.AddDays(1) };
            await _repository.UpdateBySKUIdAsync("PWTL001", updated);
            await _context.SaveChangesAsync();

            var result = await _context.Tool.FindAsync("WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6");
            Assert.Equal(updated.UnloadedDate, result!.UnloadedDate);
        }

        [Fact]
        public async Task UpdateBySKUIdAsync_DoesNothing_WhenSkuNotFound()
        {
            var item = new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow };
            _context.Tool.Add(item);
            await _context.SaveChangesAsync();

            await _repository.UpdateBySKUIdAsync("PWTL999", new Tool { PartitionKey = "WH001-PWTL999-ffffffffffffffffffffffffffffffff", SKUMarker = "PWTL999", UnloadedDate = DateTime.UtcNow.AddDays(1) });
            await _context.SaveChangesAsync();

            var result = await _context.Tool.FindAsync("WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6");
            Assert.Equal(item.UnloadedDate, result!.UnloadedDate);
        }

        [Fact]
        public async Task DeleteBySKUIdAsync_ReturnsFalse_WhenNotFound()
        {
            var result = await _repository.DeleteBySKUIdAsync("PWTL999");

            Assert.False(result);
        }

        [Fact]
        public async Task DeleteBySKUIdAsync_ReturnsTrue_AndRemovesAllMatchingItems()
        {
            _context.Tool.AddRange(
                new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow },
                new Tool { PartitionKey = "WH001-PWTL001-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow },
                new Tool { PartitionKey = "WH001-PWTL002-c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8", SKUMarker = "PWTL002", UnloadedDate = DateTime.UtcNow }
            );
            await _context.SaveChangesAsync();

            var result = await _repository.DeleteBySKUIdAsync("PWTL001");
            await _context.SaveChangesAsync();

            Assert.True(result);
            Assert.Equal(1, await _context.Tool.CountAsync());
        }

        [Fact]
        public async Task GetByLocationAsync_ReturnsMatchingItems_WhenFound()
        {
            _context.Tool.AddRange(
                new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", LocationId = "WH001", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow },
                new Tool { PartitionKey = "WH002-PWTL001-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", LocationId = "WH002", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow }
            );
            await _context.SaveChangesAsync();

            var result = await _repository.GetByLocationAsync("WH001");

            var single = Assert.Single(result);
            Assert.Equal("WH001", single.LocationId);
        }

        [Fact]
        public async Task GetByLocationAsync_ReturnsEmptyList_WhenNotFound()
        {
            var result = await _repository.GetByLocationAsync("WH999");

            Assert.Empty(result);
        }

        [Fact]
        public async Task GetByLocationAndSKUAsync_ReturnsMatchingItems_WhenFound()
        {
            _context.Tool.AddRange(
                new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", LocationId = "WH001", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow },
                new Tool { PartitionKey = "WH001-PWTL002-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", LocationId = "WH001", SKUMarker = "PWTL002", UnloadedDate = DateTime.UtcNow },
                new Tool { PartitionKey = "WH002-PWTL001-c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8", LocationId = "WH002", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow }
            );
            await _context.SaveChangesAsync();

            var result = await _repository.GetByLocationAndSKUAsync("WH001", "PWTL001");

            var single = Assert.Single(result);
            Assert.Equal("WH001", single.LocationId);
            Assert.Equal("PWTL001", single.SKUMarker);
        }

        [Fact]
        public async Task GetByLocationAndSKUAsync_ReturnsEmptyList_WhenNotFound()
        {
            var result = await _repository.GetByLocationAndSKUAsync("WH999", "PWTL999");

            Assert.Empty(result);
        }

        [Fact]
        public async Task PatchAsync_UpdatesProjectedAndUnloadedDate_WhenFound()
        {
            var item = new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow, Projected = false };
            _context.Tool.Add(item);
            await _context.SaveChangesAsync();

            var newDate = DateTime.UtcNow.AddDays(3);
            await _repository.PatchAsync("WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", true, newDate);
            await _context.SaveChangesAsync();

            var result = await _context.Tool.FindAsync("WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6");
            Assert.True(result!.Projected);
            Assert.Equal(newDate, result.UnloadedDate);
        }

        [Fact]
        public async Task PatchAsync_DoesNothing_WhenNotFound()
        {
            await _repository.PatchAsync("WH999-PWTL999-ffffffffffffffffffffffffffffffff", true, DateTime.UtcNow);
            await _context.SaveChangesAsync();

            Assert.Equal(0, await _context.Tool.CountAsync());
        }

        [Fact]
        public async Task ReceiveDeliveryAsync_MarksProjectedItemsAsReceived_UpToQuantity()
        {
            _context.Tool.AddRange(
                new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow, Projected = true },
                new Tool { PartitionKey = "WH001-PWTL001-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow, Projected = true },
                new Tool { PartitionKey = "WH001-PWTL001-c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow, Projected = true }
            );
            await _context.SaveChangesAsync();

            await _repository.ReceiveDeliveryAsync("PWTL001", 2, "WH002");
            await _context.SaveChangesAsync();

            var items = await _context.Tool.Where(t => t.SKUMarker == "PWTL001").ToListAsync();
            Assert.Equal(2, items.Count(t => !t.Projected && t.LocationId == "WH002"));
            Assert.Equal(1, items.Count(t => t.Projected));
        }

        [Fact]
        public async Task DeleteByPartitionKeyAsync_ReturnsFalse_WhenNotFound()
        {
            var result = await _repository.DeleteByPartitionKeyAsync("WH999-PWTL999-ffffffffffffffffffffffffffffffff");

            Assert.False(result);
        }

        [Fact]
        public async Task DeleteByPartitionKeyAsync_ReturnsTrue_AndRemovesItem_WhenFound()
        {
            var item = new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow };
            _context.Tool.Add(item);
            await _context.SaveChangesAsync();

            var result = await _repository.DeleteByPartitionKeyAsync("WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6");
            await _context.SaveChangesAsync();

            Assert.True(result);
            Assert.Equal(0, await _context.Tool.CountAsync());
        }
    }
}
