import { useEffect, useRef } from "react";

interface Params {
    path: string;

    onMessage: (message: WSMessage) => void;
    onError?: () => void;
}

export interface WSMessage {
    message_type: number;
    content?: string;
}

export const PingMessage = 1
export const PongMessage = 1
export const CompletitionsMessage = 2
export const CancelMessage = 3

export const CompletitionsStart = 2
export const CompletitionsNext = 3
export const CompletitionsEnd = 4
export const CompletitionsQueue = 5

export function useWebSocket({
    path,
    onMessage,
    onError,
}: Params) {
    const socket = useRef<WebSocket | null>(null);
    const pingTimeout = useRef<number>(0);
    const reconnectInterval = useRef<number>(0);

    function newConnection() {
        socket.current = new WebSocket(path)
        socket.current.onopen = function (_) {
            console.info("ws opened...");
            clearInterval(reconnectInterval.current)

            // starting ping/pong messaging
            socket.current!.send(JSON.stringify({
                message_type: PingMessage
            }));
        }
        socket.current.onclose = function () {
            console.info("ws closed...");
            
            clearTimeout(pingTimeout.current)
            socket.current = null;

            clearInterval(reconnectInterval.current)
            reconnectInterval.current = setInterval(() => {
                if(!socket.current || socket.current.CLOSED) {
                    newConnection()
                }
            }, 1000)
        }
        socket.current.onmessage = function (e) {
            const message = JSON.parse(e.data) as WSMessage;
            console.log("ws got", e.data)
            if (message.message_type == PongMessage) {
                pingTimeout.current = setTimeout(() => {
                    socket.current!.send(JSON.stringify({
                        message_type: PingMessage
                    }));
                }, 1000);
                return;
            }

            onMessage(message);
        }
        socket.current.onerror = function (_) {
            console.error("ws error...");
            clearInterval(reconnectInterval.current)

            if (onError) onError()
        }
    }

    useEffect(() => {
        if (socket.current != null) {
            return;
        }

        window.onload = function () {
            newConnection()
        }

        return () => {
            clearInterval(reconnectInterval.current)
            if (socket.current == null) {
                return;
            }

            socket.current.close();
        }
    }, []);

    return {
        send(message: string) {
            if (socket.current) {
                socket.current.send(message);
            }
        }
    }
}
