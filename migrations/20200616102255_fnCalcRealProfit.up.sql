create or alter function dbo.fnCalcRealProfit(@ABetPrice decimal(9, 5), @BBetPrice decimal(9, 5)) returns decimal(9, 3) as
begin
    if @ABetPrice = 0 or @BBetPrice = 0
        return 0
    return 1/(1/@ABetPrice+1/@BBetPrice)*100-100
end