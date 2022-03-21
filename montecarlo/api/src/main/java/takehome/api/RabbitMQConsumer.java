package takehome.api;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.rabbitmq.client.Channel;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
import org.springframework.amqp.support.AmqpHeaders;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.sql.Timestamp;

@Component
public class RabbitMQConsumer {
	private static final Logger log = LoggerFactory.getLogger(RabbitMQConsumer.class);

	private final ObjectMapper objectMapper;
	private final PricesRepository repository;

	public RabbitMQConsumer(ObjectMapper objectMapper, PricesRepository repository) {
		this.objectMapper = objectMapper;
		this.repository = repository;
	}

	@RabbitListener(queues = "${montecarlo.consumer.queue}")
	public void receivedMessage(String payload, Channel channel, @Header(AmqpHeaders.DELIVERY_TAG) long tag) {
		log.info("Received Message From RabbitMQ: " + payload);
		PriceMessage message;
		try {
			message = objectMapper.readValue(payload, PriceMessage.class);
			Price price = new Price(message.getInstrument(), Timestamp.from(message.getTimestamp()), message.getPrice());
			repository.save(price);
			try {
				channel.basicAck(tag, false);
			} catch (IOException ex) {
				log.error("failed to ack message", ex);
			}
		} catch (JsonProcessingException e) {
			log.error("failed to deserialize message", e);
			try {
				channel.basicNack(tag, false, true);
			} catch (IOException ex) {
				log.error("failed to nack and requeue message", ex);
			}
		}
	}
}
