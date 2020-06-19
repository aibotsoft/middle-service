create or alter view dbo.vByMonth as
select top 1000 DATEPART(Year, sysdatetimeoffset())                        year,
                DATEPART(Month, sysdatetimeoffset())                       month,
                count(distinct b.SurebetId)                                count,
                cast(sum(b.WinLoss) as int)                                gross,
                cast(sum(b.Stake) as int)                                  depo,
                cast(sum(b.WinLoss) * 100 / sum(b.Stake) as decimal(9, 2)) perc
from dbo.BetList b
         join dbo.Side s on s.SurebetId = b.SurebetId and s.ToBetId = b.BetId
group by DATEPART(Year, s.CreatedAt), DATEPART(Month, s.CreatedAt)
order by 1 desc;