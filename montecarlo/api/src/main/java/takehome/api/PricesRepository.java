package takehome.api;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface PricesRepository extends JpaRepository<Price, UUID> {
    @Query(nativeQuery = true, value = "select distinct instrument from prices order by instrument")
    List<String> allMarkets();

    @Query(nativeQuery = true, value = """
            select instrument,
                   rank() over (order by stddev(amount) desc) || '/' || count(*) over () as rank
            from prices
            group by instrument
            """)
    List<Market> marketRanks();

    @Query(nativeQuery = true, value = """
            select prices.timestamp, prices.amount
            from prices
            where timestamp >= now() - interval '24 hours'
              and instrument = ?1
            """)
    List<PricePoint> getPricePoints(String instrument);
}
