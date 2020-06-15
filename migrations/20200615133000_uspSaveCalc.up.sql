create or alter proc dbo.uspSaveCalc @Profit decimal(9, 5),
                                     @FirstName varchar(1000),
                                     @SecondName varchar(1000),
                                     @LowerWinIndex tinyint,
                                     @HigherWinIndex tinyint,
                                     @FirstIndex tinyint,
                                     @SecondIndex tinyint,
                                     @WinDiff decimal(9, 5),
                                     @WinDiffRel decimal(9, 5),
                                     @FortedSurebetId bigint,
                                     @SurebetId bigint
as
begin
    set nocount on
    declare @Id bigint

    select @Id = SurebetId from dbo.Calc where SurebetId = @SurebetId
    if @@rowcount = 0
        insert into dbo.Calc(Profit, FirstName, SecondName, LowerWinIndex, HigherWinIndex, FirstIndex, SecondIndex,
                             WinDiff, WinDiffRel, FortedSurebetId, SurebetId)
        output inserted.SurebetId
        values (@Profit, @FirstName, @SecondName, @LowerWinIndex, @HigherWinIndex, @FirstIndex, @SecondIndex, @WinDiff,
                @WinDiffRel, @FortedSurebetId, @SurebetId)
    else
        select @Id
end