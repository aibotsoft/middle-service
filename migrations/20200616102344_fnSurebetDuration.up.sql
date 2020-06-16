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