FROM maven:3-openjdk-17-slim AS build
WORKDIR /usr/src/app
COPY src src
COPY pom.xml pom.xml
RUN mvn clean package

FROM gcr.io/distroless/java17
WORKDIR /usr/app
COPY --from=build /usr/src/app/target/*.jar app.jar
CMD ["app.jar"]
