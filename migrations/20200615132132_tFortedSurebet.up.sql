create table dbo.FortedSurebet
(
    CreatedAt       datetimeoffset,
    Starts          datetimeoffset,
    FortedHome      varchar(1000),
    FortedAway      varchar(1000),
    FortedProfit    varchar(1000),
    FortedSport     varchar(1000),
    FortedLeague    varchar(1000),
    FilterName      varchar(1000),
    FortedSurebetId bigint not null,
    constraint PK_FortedSurebet primary key (FortedSurebetId),
)
