package takehome.originator;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.jetbrains.annotations.NotNull;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.MalformedURLException;
import java.net.ProtocolException;
import java.net.URL;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;

@Component
public class GetLatestPrices {
    private static class Allowance {
        private double cost;
        private double remaining;
        private String upgrade;

        public double getCost() {
            return cost;
        }

        public void setCost(double cost) {
            this.cost = cost;
        }

        public double getRemaining() {
            return remaining;
        }

        public void setRemaining(double remaining) {
            this.remaining = remaining;
        }

        public String getUpgrade() {
            return upgrade;
        }

        public void setUpgrade(String upgrade) {
            this.upgrade = upgrade;
        }
    }

    private static class CryptoWatchMarkets {
        private List<Market> result;
        private Allowance allowance;

        public Allowance getAllowance() {
            return allowance;
        }

        public void setAllowance(Allowance allowance) {
            this.allowance = allowance;
        }

        public List<Market> getResult() {
            return result;
        }

        public void setResult(List<Market> result) {
            this.result = result;
        }
    }

    private static class CryptoWatchPrice {
        private Price result;
        private Allowance allowance;

        public Price getResult() {
            return result;
        }

        public void setResult(Price result) {
            this.result = result;
        }

        public Allowance getAllowance() {
            return allowance;
        }

        public void setAllowance(Allowance allowance) {
            this.allowance = allowance;
        }
    }

    private static final Logger log = LoggerFactory.getLogger(GetLatestPrices.class);

    private final ObjectMapper objectMapper;
    private final RabbitTemplate rabbitTemplate;

    @Value("${montecarlo.producer.exchange}")
    private String exchange;

    @Value("${montecarlo.producer.routingKey}")
    private String routingKey;

    @Value("${cryptowatch.markets.url}")
    private String marketURL;

    public GetLatestPrices(ObjectMapper objectMapper, RabbitTemplate rabbitTemplate) {
        this.objectMapper = objectMapper;
        this.rabbitTemplate = rabbitTemplate;
    }

    @Scheduled(fixedRate = 1, timeUnit = TimeUnit.MINUTES)
    public void run() {
        List<Market> activeMarkets = fetchMarkets().stream().filter(Market::getActive).toList();
        for (Market market : activeMarkets) {
            Price price = fetchPrice(market);
            String message = String.format("{\"instrument\": \"%s\", \"price\":\"%s\", \"timestamp\":\"%s\"}", market.getPair(), price.getPrice(), Instant.now().toString());
            log.debug("Publishing price: " + message);
            rabbitTemplate.convertAndSend(exchange, routingKey, message);
        }
    }

    public Price fetchPrice(@NotNull Market market) {
        Price price = null;
        URL url;
        HttpURLConnection con = null;
        String priceURL = market.getRoute() + "/price";
        try {
            url = new URL(priceURL);
            con = (HttpURLConnection) url.openConnection();
            con.setRequestMethod("GET");
            int status = con.getResponseCode();
            if (status != 200) {
                log.error("received unexpected status code: {}", status);
                return null;
            }
            BufferedReader in = new BufferedReader(new InputStreamReader(con.getInputStream()));
            String inputLine;
            StringBuilder content = new StringBuilder();
            while ((inputLine = in.readLine()) != null) {
                content.append(inputLine);
            }
            in.close();
            CryptoWatchPrice parsedPrice = objectMapper.readValue(content.toString(), CryptoWatchPrice.class);
            price = parsedPrice.getResult();
        } catch (MalformedURLException e) {
            log.error("failed to parse market price URL: " + priceURL, e);
        } catch (ProtocolException e) {
            log.error("failed to open connection to " + priceURL, e);
        } catch (IOException e) {
            log.error("failed to set request method", e);
        } finally {
            if (con != null) {
                con.disconnect();
            }
        }
        return price;
    }

    public List<Market> fetchMarkets() {
        List<Market> markets = new ArrayList<>();
        URL url;
        HttpURLConnection con = null;
        try {
            url = new URL(marketURL);
            con = (HttpURLConnection) url.openConnection();
            con.setRequestMethod("GET");
            int status = con.getResponseCode();
            if (status != 200) {
                log.error("received unexpected status code: {}", status);
                return markets;
            }
            BufferedReader in = new BufferedReader(new InputStreamReader(con.getInputStream()));
            String inputLine;
            StringBuilder content = new StringBuilder();
            while ((inputLine = in.readLine()) != null) {
                content.append(inputLine);
            }
            in.close();
            CryptoWatchMarkets parsedMarkets = objectMapper.readValue(content.toString(), CryptoWatchMarkets.class);
            markets = parsedMarkets.getResult();
        } catch (MalformedURLException e) {
            log.error("failed to parse markets URL: " + marketURL, e);
        } catch (ProtocolException e) {
            log.error("failed to open connection to " + marketURL, e);
        } catch (IOException e) {
            log.error("failed to set request method", e);
        } finally {
            if (con != null) {
                con.disconnect();
            }
        }
        return markets;
    }
}
