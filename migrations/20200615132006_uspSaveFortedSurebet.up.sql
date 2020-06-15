create or alter proc dbo.uspSaveFortedSurebet @CreatedAt datetimeoffset,
                                              @Starts datetimeoffset,
                                              @FortedHome varchar(1000),
                                              @FortedAway varchar(1000),
                                              @FortedProfit varchar(1000),
                                              @FortedSport varchar(1000),
                                              @FortedLeague varchar(1000),
                                              @FilterName varchar(1000),
                                              @FortedSurebetId bigint
as
begin
    set nocount on
    declare @Id bigint

    select @Id = FortedSurebetId from dbo.FortedSurebet where FortedSurebetId = @FortedSurebetId
    if @@rowcount = 0
        insert into dbo.FortedSurebet(CreatedAt, Starts, FortedHome, FortedAway, FortedProfit, FortedSport,
                                      FortedLeague, FilterName, FortedSurebetId)
        output inserted.FortedSurebetId

        values (@CreatedAt, @Starts, @FortedHome, @FortedAway, @FortedProfit, @FortedSport, @FortedLeague, @FilterName,
                @FortedSurebetId)
    else
        select @Id
end