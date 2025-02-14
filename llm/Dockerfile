
FROM golang:1.23.6-bookworm AS build

RUN apt-get update && \
    apt-get install -y build-essential git cmake libcurl4-openssl-dev

WORKDIR /app

COPY . .

RUN cmake -S ./deps/llama.cpp -B ./deps/llama.cpp/build -DCMAKE_BUILD_TYPE=Release \
    && cmake --build ./deps/llama.cpp/build --config Release -j $(nproc)      

RUN mkdir -p /app/lib && \
    find deps/llama.cpp/build -name "*.so" -exec cp {} /app/lib \;

RUN go build -C ./cmd -o llm

FROM ubuntu:22.04

RUN apt-get update \
    && apt-get install -y libgomp1 curl\
    && apt autoremove -y \
    && apt clean -y \
    && rm -rf /tmp/* /var/tmp/* \
    && find /var/cache/apt/archives /var/lib/apt/lists -not -name lock -type f -delete \
    && find /var/cache -type f -delete

COPY --from=build /app/lib/ /app
COPY --from=build /app/cmd/llm .
ENV LD_LIBRARY_PATH=/app:$LD_LIBRARY_PATH

CMD ["./llm"]
