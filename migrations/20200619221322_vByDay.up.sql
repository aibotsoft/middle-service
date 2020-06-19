create or alter view dbo.vByDay as
select top 1000 cast(s.CreatedAt AS date)                                  day,
                count(distinct b.SurebetId)                                count,
                cast(sum(b.WinLoss) as int)                                gross,
                cast(sum(b.Stake) as int)                                  depo,
                cast(sum(b.WinLoss) * 100 / sum(b.Stake) as decimal(9, 2)) perc
from dbo.BetList b
         join dbo.Side s on s.SurebetId = b.SurebetId and s.ToBetId = b.BetId
group by CAST(s.CreatedAt AS date)
order by 1 desc;


