using Xunit;
using Microsoft.EntityFrameworkCore;
using Switchyard.LogisticsAPI.Data;
using Switchyard.Domain;

namespace Switchyard.LogisticsAPI.Tests.Data
{
    public class UnitOfWorkTests : IDisposable
    {
        private readonly LogisticsContext _context;
        private readonly LogisticsReadContext _readContext;
        private readonly UnitOfWork _unitOfWork;

        public UnitOfWorkTests()
        {
            var dbName = Guid.NewGuid().ToString();
            var writeOptions = new DbContextOptionsBuilder<LogisticsContext>()
                .UseInMemoryDatabase(dbName)
                .Options;
            var readOptions = new DbContextOptionsBuilder<LogisticsReadContext>()
                .UseInMemoryDatabase(dbName)
                .Options;
            _context = new LogisticsContext(writeOptions);
            _context.Database.EnsureCreated();
            _readContext = new LogisticsReadContext(readOptions);
            _unitOfWork = new UnitOfWork(_context, _readContext);
        }

        public void Dispose() => _unitOfWork.Dispose();

        [Fact]
        public void BillsOfLading_LineEntries_Warehouses_Stores_ReturnRepositoryInstances_AndAreMemoized()
        {
            var bol1 = _unitOfWork.BillsOfLading;
            var bol2 = _unitOfWork.BillsOfLading;
            var le1 = _unitOfWork.LineEntries;
            var le2 = _unitOfWork.LineEntries;
            var wh1 = _unitOfWork.Warehouses;
            var wh2 = _unitOfWork.Warehouses;
            var st1 = _unitOfWork.Stores;
            var st2 = _unitOfWork.Stores;

            Assert.NotNull(bol1);
            Assert.Same(bol1, bol2);
            Assert.NotNull(le1);
            Assert.Same(le1, le2);
            Assert.NotNull(wh1);
            Assert.Same(wh1, wh2);
            Assert.NotNull(st1);
            Assert.Same(st1, st2);
        }

        [Fact]
        public async Task SaveChangesAsync_PersistsPendingChanges()
        {
            _context.Warehouses.Add(new Warehouse { WarehouseId = "WH001", City = "Springfield", State = "IL" });

            var result = await _unitOfWork.SaveChangesAsync();

            Assert.Equal(1, result);
            Assert.Equal(1, await _context.Warehouses.CountAsync());
        }
    }
}
