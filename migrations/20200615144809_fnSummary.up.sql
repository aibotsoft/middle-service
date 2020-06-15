create or alter function dbo.fnCalcRealProfit(@ABetPrice decimal(9, 5), @BBetPrice decimal(9, 5)) returns decimal(9, 3) as
begin
    if @ABetPrice = 0 or @BBetPrice = 0
        return 0
    return 1/(1/@ABetPrice+1/@BBetPrice)*100-100
end
go

create or alter function dbo.fnSurebetDuration(@SurebetId bigint, @ADone bigint, @BDone bigint) returns int
    WITH SCHEMABINDING as
begin
    if @ADone = 0 and @BDone = 0
        return 0
    if @ADone > @BDone
        return @ADone - @SurebetId / 1000
    else
        return @BDone - @SurebetId / 1000
    return 0
end
go

create or alter function dbo.fnWinLossStatusCheck(@WinLoss decimal(9, 5), @Status varchar(1000)) returns decimal(9, 5)
    WITH SCHEMABINDING as
begin
    if @Status in ('Running', 'ACCEPTED', '')
        return null
    return @WinLoss
end
go

create or alter function dbo.fnCalcProfit(@AWinLoss decimal(9, 5), @BWinLoss decimal(9, 5)) returns decimal(9, 5)
    WITH SCHEMABINDING as
begin
    if @AWinLoss is not null and @BWinLoss is not null
        return @AWinLoss + @BWinLoss
    return null
end
go