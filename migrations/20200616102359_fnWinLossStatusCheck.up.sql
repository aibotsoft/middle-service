create or alter function dbo.fnWinLossStatusCheck(@WinLoss decimal(9, 5), @Status varchar(1000)) returns decimal(9, 5)
    WITH SCHEMABINDING as
begin
    if @Status in ('Running', 'ACCEPTED', '')
        return null
    return @WinLoss
end