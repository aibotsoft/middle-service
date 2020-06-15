create or alter view dbo.Summary as
select c.CreatedAt,
       f.Starts,
       c.Profit,
       dbo.fnCalcRealProfit(a.BetPrice, b.BetPrice)                            realProfit,
       dbo.fnSurebetDuration(c.SurebetId, a.CheckDone, b.CheckDone)            check_time,
       dbo.fnSurebetDuration(c.SurebetId, a.Done, b.Done)                      bet_time,

       dbo.fnCalcProfit(dbo.fnWinLossStatusCheck(al.WinLoss, al.ApiBetStatus),
                        dbo.fnWinLossStatusCheck(bl.WinLoss, bl.ApiBetStatus)) win,
       dbo.fnWinLossStatusCheck(al.WinLoss, al.ApiBetStatus)                   a_win_loss,
       dbo.fnWinLossStatusCheck(bl.WinLoss, bl.ApiBetStatus)                   b_win_loss,
--        al.WinLoss                                                   a_win_loss,
--        bl.WinLoss                                                   b_win_loss,
       a.ServiceName                                                           a_service,
       b.ServiceName                                                           b_service,
       a.BetStake                                                              a_stake,
       b.BetStake                                                              b_stake,
       a.BetPrice                                                              a_price,
       b.BetPrice                                                              b_price,
       a.MarketName                                                            a_market,
       b.MarketName                                                            b_market,
       a.BetStatus                                                             a_bet_status,
       b.BetStatus                                                             b_bet_status,
       a.BetStatusInfo                                                         a_info,
       b.BetStatusInfo                                                         b_info,
       a.ApiBetId                                                              a_api_bet,
       b.ApiBetId                                                              b_api_bet,
       f.FortedSport                                                           sport,
       f.FortedLeague                                                          league,
       f.FortedHome                                                            forted_home,
       f.FortedAway                                                            forted_away,
       f.FortedSurebetId                                                       forted_id,
       a.CheckDone - a.CheckId                                                 a_check,
       b.CheckDone - b.CheckId                                                 b_check,
       a.MaxBet                                                                a_max_bet,
       b.MaxBet                                                                b_max_bet,

       a.TryCount                                                              a_try,
       b.TryCount                                                              b_try,
       c.SurebetId
from dbo.Calc c
         join dbo.FortedSurebet f on f.FortedSurebetId = c.FortedSurebetId
         join dbo.Side a on a.SurebetId = c.SurebetId and a.SideIndex = 0
         join dbo.Side b on b.SurebetId = c.SurebetId and b.SideIndex = 1
         left join dbo.BetList al on al.SurebetId = a.SurebetId and al.BetId = a.ToBetId
         left join dbo.BetList bl on bl.SurebetId = b.SurebetId and bl.BetId = b.ToBetId

create or alter function dbo.fnCalcRealProfit(@ABetPrice decimal(9, 5), @BBetPrice decimal(9, 5)) returns decimal(9, 3) as
begin
    if @ABetPrice = 0 or @BBetPrice = 0
        return 0
    return 1/(1/@ABetPrice+1/@BBetPrice)*100-100
end


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

create or alter function dbo.fnWinLossStatusCheck(@WinLoss decimal(9, 5), @Status varchar(1000)) returns decimal(9, 5)
    WITH SCHEMABINDING as
begin
    if @Status in ('Running', 'ACCEPTED', '')
        return null
    return @WinLoss
end

create or alter function dbo.fnCalcProfit(@AWinLoss decimal(9, 5), @BWinLoss decimal(9, 5)) returns decimal(9, 5)
    WITH SCHEMABINDING as
begin
    if @AWinLoss is not null and @BWinLoss is not null
        return @AWinLoss + @BWinLoss
    return null
end
