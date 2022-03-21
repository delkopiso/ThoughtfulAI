package takehome.api;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@RestController
public class MetricsController {
    PricesRepository repository;

    @Autowired
    public MetricsController(PricesRepository repository) {
        this.repository = repository;
    }

    @GetMapping("/metrics")
    List<String> all() {
        return this.repository.allMarkets();
    }

    @GetMapping("/metrics/{name}")
    Map<String, Object> one(@PathVariable String name) {
        Map<String, String> markets = repository.marketRanks().stream().collect(
                Collectors.toMap(Market::getInstrument, Market::getRank)
        );
        Map<String, Object> response = new HashMap<>();
        response.put("rank", markets.get(name));
        response.put("pricePoints", repository.getPricePoints(name));
        return response;
    }

}
