create table dbo.Calc
(
    SurebetId        bigint                                        not null,
    FortedSurebetId  bigint                                        not null,
    Profit           decimal(9, 5),

    EffectiveProfit  decimal(9, 5),
    MiddleDiff       decimal(9, 5),
    MiddleMargin     decimal(9, 5),
    HoursBeforeEvent decimal(9, 5),
    Gross            decimal(9, 5),
    SurebetType      varchar(1000),

    FirstName        varchar(1000),
    SecondName       varchar(1000),
    LowerWinIndex    tinyint,
    HigherWinIndex   tinyint,
    FirstIndex       tinyint,
    SecondIndex      tinyint,
    WinDiff          decimal(9, 5),
    WinDiffRel       decimal(9, 5),

    CreatedAt        datetimeoffset(2) default sysdatetimeoffset() not null,
    constraint PK_Calc primary key (SurebetId),
)