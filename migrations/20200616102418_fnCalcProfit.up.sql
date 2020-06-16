create or alter function dbo.fnCalcProfit(@AWinLoss decimal(9, 5), @BWinLoss decimal(9, 5)) returns decimal(9, 5)
    WITH SCHEMABINDING as
begin
    if @AWinLoss is not null and @BWinLoss is not null
        return @AWinLoss + @BWinLoss
    return null
end