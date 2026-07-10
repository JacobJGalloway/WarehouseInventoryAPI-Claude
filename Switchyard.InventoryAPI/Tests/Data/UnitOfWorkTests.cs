using Xunit;
using Microsoft.EntityFrameworkCore;
using Switchyard.InventoryAPI.Data;
using Switchyard.Domain;

namespace Switchyard.InventoryAPI.Tests.Data
{
    public class UnitOfWorkTests : IDisposable
    {
        private readonly InventoryContext _context;
        private readonly InventoryReadContext _readContext;
        private readonly UnitOfWork _unitOfWork;

        public UnitOfWorkTests()
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
            _unitOfWork = new UnitOfWork(_context, _readContext);
        }

        public void Dispose()
        {
            _unitOfWork.Dispose();
        }

        [Fact]
        public void Clothing_PPE_Tools_ReturnRepositoryInstances_AndAreMemoized()
        {
            var clothing1 = _unitOfWork.Clothing;
            var clothing2 = _unitOfWork.Clothing;
            var ppe1 = _unitOfWork.PPE;
            var ppe2 = _unitOfWork.PPE;
            var tools1 = _unitOfWork.Tools;
            var tools2 = _unitOfWork.Tools;

            Assert.NotNull(clothing1);
            Assert.Same(clothing1, clothing2);
            Assert.NotNull(ppe1);
            Assert.Same(ppe1, ppe2);
            Assert.NotNull(tools1);
            Assert.Same(tools1, tools2);
        }

        [Fact]
        public async Task GetClothingBySKUIdAsync_ReturnsMatchingItems()
        {
            _context.Clothing.Add(new Clothing { PartitionKey = "WH001-CLTH001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });
            await _context.SaveChangesAsync();

            var result = await _unitOfWork.GetClothingBySKUIdAsync("CLTH001");

            var single = Assert.Single(result);
            Assert.Equal("CLTH001", single.SKUMarker);
        }

        [Fact]
        public async Task GetPPEBySKUIdAsync_ReturnsMatchingItems()
        {
            _context.PPE.Add(new PPE { PartitionKey = "WH001-SPPE001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "SPPE001", UnloadedDate = DateTime.UtcNow });
            await _context.SaveChangesAsync();

            var result = await _unitOfWork.GetPPEBySKUIdAsync("SPPE001");

            var single = Assert.Single(result);
            Assert.Equal("SPPE001", single.SKUMarker);
        }

        [Fact]
        public async Task GetToolBySKUIdAsync_ReturnsMatchingItems()
        {
            _context.Tool.Add(new Tool { PartitionKey = "WH001-PWTL001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "PWTL001", UnloadedDate = DateTime.UtcNow });
            await _context.SaveChangesAsync();

            var result = await _unitOfWork.GetToolBySKUIdAsync("PWTL001");

            var single = Assert.Single(result);
            Assert.Equal("PWTL001", single.SKUMarker);
        }

        [Fact]
        public async Task ReceiveDeliveryAsync_MarksMatchingProjectedItems_AcrossAllCategories_AndSaves()
        {
            _context.Clothing.Add(new Clothing { PartitionKey = "WH001-SHARED001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", SKUMarker = "SHARED001", UnloadedDate = DateTime.UtcNow, Projected = true });
            _context.PPE.Add(new PPE { PartitionKey = "WH001-SHARED001-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7", SKUMarker = "SHARED001", UnloadedDate = DateTime.UtcNow, Projected = true });
            _context.Tool.Add(new Tool { PartitionKey = "WH001-SHARED001-c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8", SKUMarker = "SHARED001", UnloadedDate = DateTime.UtcNow, Projected = true });
            await _context.SaveChangesAsync();

            await _unitOfWork.ReceiveDeliveryAsync("WH002", [new DeliveryLineItem { SKUMarker = "SHARED001", Quantity = 1 }]);

            var clothing = await _context.Clothing.FindAsync("WH001-SHARED001-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6");
            var ppe = await _context.PPE.FindAsync("WH001-SHARED001-b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7");
            var tool = await _context.Tool.FindAsync("WH001-SHARED001-c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8");

            Assert.False(clothing!.Projected);
            Assert.Equal("WH002", clothing.LocationId);
            Assert.False(ppe!.Projected);
            Assert.Equal("WH002", ppe.LocationId);
            Assert.False(tool!.Projected);
            Assert.Equal("WH002", tool.LocationId);
        }

        [Fact]
        public async Task SaveChangesAsync_PersistsPendingChanges()
        {
            _context.Clothing.Add(new Clothing { SKUMarker = "CLTH001", UnloadedDate = DateTime.UtcNow });

            var result = await _unitOfWork.SaveChangesAsync();

            Assert.Equal(1, result);
            Assert.Equal(1, await _context.Clothing.CountAsync());
        }
    }
}
