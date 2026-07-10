using Xunit;
using Microsoft.EntityFrameworkCore;
using Switchyard.InventoryAPI.Data;
using Switchyard.Domain;

namespace Switchyard.InventoryAPI.Tests.Data
{
    public class InventoryContextTests : IDisposable
    {
        private readonly InventoryContext _context;

        public InventoryContextTests()
        {
            var options = new DbContextOptionsBuilder<InventoryContext>()
                .UseInMemoryDatabase(Guid.NewGuid().ToString())
                .Options;
            _context = new InventoryContext(options);
            _context.Database.EnsureCreated();
        }

        public void Dispose() => _context.Dispose();

        [Fact]
        public async Task GetClothingBySKUIdsync_ReturnsMatchingItems_WhenFound()
        {
            _context.Clothing.Add(new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });
            await _context.SaveChangesAsync();

            var result = await _context.GetClothingBySKUIdsync("CLTH001");

            var single = Assert.Single(result);
            Assert.Equal("CLTH001", single.SKUMarker);
        }

        [Fact]
        public async Task GetClothingBySKUIdsync_ReturnsEmptyList_WhenNotFound()
        {
            var result = await _context.GetClothingBySKUIdsync("CLTH999");

            Assert.Empty(result);
        }

        [Fact]
        public async Task GetPPEBySKUIdsync_ReturnsMatchingItems_WhenFound()
        {
            _context.PPE.Add(new PPE { PartitionKey = "WH001-SPPE001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "SPPE001", UnloadedDate = DateTime.UtcNow });
            await _context.SaveChangesAsync();

            var result = await _context.GetPPEBySKUIdsync("SPPE001");

            var single = Assert.Single(result);
            Assert.Equal("SPPE001", single.SKUMarker);
        }

        [Fact]
        public async Task GetPPEBySKUIdsync_ReturnsEmptyList_WhenNotFound()
        {
            var result = await _context.GetPPEBySKUIdsync("SPPE999");

            Assert.Empty(result);
        }

        [Fact]
        public async Task GetToolBySKUIdsync_ReturnsMatchingItems_WhenFound()
        {
            _context.Tool.Add(new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow });
            await _context.SaveChangesAsync();

            var result = await _context.GetToolBySKUIdsync("PWTL001");

            var single = Assert.Single(result);
            Assert.Equal("PWTL001", single.SKUMarker);
        }

        [Fact]
        public async Task GetToolBySKUIdsync_ReturnsEmptyList_WhenNotFound()
        {
            var result = await _context.GetToolBySKUIdsync("PWTL999");

            Assert.Empty(result);
        }
    }
}
