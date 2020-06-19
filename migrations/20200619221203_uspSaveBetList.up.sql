create or alter proc dbo.uspSaveBetList @TVP dbo.BetListType READONLY as
begin
    set nocount on

    MERGE dbo.BetList AS t
    USING @TVP s
    ON (t.SurebetId = s.SurebetId and t.BetId = s.BetId)

    WHEN MATCHED THEN
        UPDATE
        SET Price    = s.Price,
            Stake     = s.Stake,
            WinLoss   = s.WinLoss,
            ApiBetId    = s.ApiBetId,
            ApiBetStatus      = s.ApiBetStatus,
            UpdatedAt =sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (SurebetId, BetId, Price, Stake, WinLoss, ApiBetId, ApiBetStatus)
        VALUES (s.SurebetId, s.BetId, s.Price, s.Stake, s.WinLoss, s.ApiBetId, s.ApiBetStatus);
end