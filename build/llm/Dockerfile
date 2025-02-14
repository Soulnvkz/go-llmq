
FROM golang:1.23.6-bookworm AS build

RUN apt-get update && \
    apt-get install -y build-essential git cmake libcurl4-openssl-dev

WORKDIR /project

COPY ./llm ./llm
COPY ./pkg ./pkg

RUN cmake -S ./llm/deps/llama.cpp -B ./llm/deps/llama.cpp/build -DCMAKE_BUILD_TYPE=Release \
    && cmake --build ./llm/deps/llama.cpp/build --config Release -j $(nproc)      

RUN mkdir -p /project/lib && \
    find llm/deps/llama.cpp/build -name "*.so" -exec cp {} /project/lib \;

RUN go mod tidy -C ./llm
RUN go build -C ./llm/cmd -o llm

FROM ubuntu:22.04

WORKDIR /app

RUN apt-get update \
    && apt-get install -y libgomp1 curl\
    && apt autoremove -y \
    && apt clean -y \
    && rm -rf /tmp/* /var/tmp/* \
    && find /var/cache/apt/archives /var/lib/apt/lists -not -name lock -type f -delete \
    && find /var/cache -type f -delete

COPY --from=build /project/lib/ /app
COPY --from=build /project/llm/cmd/llm .
ENV LD_LIBRARY_PATH=/app:$LD_LIBRARY_PATH

CMD ["./llm"]
