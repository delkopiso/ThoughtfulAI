package takehome.api;

import java.sql.Timestamp;

interface PricePoint {
    Timestamp getTimestamp();
    Double getAmount();
}
