create table dbo.BetConfig
(
    ServiceName    varchar(50)   not null,
    Regime         varchar(50)   not null default 'Disabled',
    MinStake       int           not null default 1,
    MaxStake       int           not null default 1,
    MaxWin         int           not null default 10,
    MinPercent     decimal(9, 5) not null default 1,
    MaxPercent     tinyint       not null default 10,
    Priority       tinyint       not null default 255,
    MaxCountLine   tinyint       not null default 1,
    MaxCountEvent  tinyint       not null default 2,
    MaxAmountEvent int           not null default 100,
    MaxAmountLine  int           not null default 100,
    RoundValue     decimal(9, 5) not null default 1,
    constraint PK_BetConfig primary key (ServiceName),
    constraint CK_BetConfig_Regime check (Regime IN ('Disabled', 'Surebet', 'OnlyCheck')),
)
insert into dbo.BetConfig (ServiceName, Regime, MinStake, MaxStake, MinPercent, MaxPercent, Priority, MaxCountLine,
                           MaxCountEvent, MaxAmountEvent,MaxAmountLine, RoundValue)
values ('Pinnacle', 'Surebet', default, default, default, default, default, default, default, default,default, 0.01),
       ('Sbobet', 'Surebet', default, default, default, default, default, default, default, default,default, default),
       ('Dafabet', 'Surebet', default, default, default, default, default, default, default, default,default, default),
       ('SportMarket', 'Surebet', default, default, default, default, default, default, default, default,default, 0.01)
