create or alter proc dbo.uspSaveCalc @Profit decimal(9, 5),
                                     @EffectiveProfit  decimal(9, 5),
                                     @MiddleDiff       decimal(9, 5),
                                     @MiddleMargin     decimal(9, 5),
                                     @HoursBeforeEvent decimal(9, 5),
                                     @Gross            decimal(9, 5),
                                     @SurebetType      varchar(1000),

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
        insert into dbo.Calc(Profit,
                             EffectiveProfit,
                             MiddleDiff,
                             MiddleMargin,
                             HoursBeforeEvent,
                             Gross,
                             SurebetType,
                             FirstName, SecondName, LowerWinIndex, HigherWinIndex, FirstIndex, SecondIndex,
                             WinDiff, WinDiffRel, FortedSurebetId, SurebetId)
        output inserted.SurebetId
        values (@Profit,
                @EffectiveProfit,
                @MiddleDiff,
                @MiddleMargin,
                @HoursBeforeEvent,
                @Gross,
                @SurebetType,

                @FirstName,
                @SecondName, @LowerWinIndex, @HigherWinIndex, @FirstIndex, @SecondIndex, @WinDiff,
                @WinDiffRel, @FortedSurebetId, @SurebetId)
    else
        select @Id
end