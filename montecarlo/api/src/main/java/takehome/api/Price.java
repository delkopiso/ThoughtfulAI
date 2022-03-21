package takehome.api;

import jakarta.persistence.*;
import org.hibernate.annotations.Type;

import java.sql.Timestamp;
import java.util.UUID;

@Entity
@Table(name = "prices", indexes = {
        @Index(columnList = "instrument"),
        @Index(columnList = "instrument, timestamp"),
})
public class Price {
    @Id
    @Column(name = "price_id", nullable = false)
    @Type(type = "pg-uuid")
    private UUID id;

    @Column(name = "instrument", nullable = false)
    private String instrument;

    @Column(name = "timestamp", nullable = false)
    private Timestamp timestamp;

    @Column(name = "amount", nullable = false)
    private double amount;

    public Price(String instrument, Timestamp timestamp, double amount) {
        super();
        this.id = UUID.randomUUID();
        this.instrument = instrument;
        this.timestamp = timestamp;
        this.amount = amount;
    }

    public Price() {
    }

    public UUID getId() {
        return id;
    }

    public void setId(UUID id) {
        this.id = id;
    }

    public String getInstrument() {
        return instrument;
    }

    public void setInstrument(String instrument) {
        this.instrument = instrument;
    }

    public Timestamp getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(Timestamp timestamp) {
        this.timestamp = timestamp;
    }

    public double getAmount() {
        return amount;
    }

    public void setAmount(double amount) {
        this.amount = amount;
    }
}
