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

    useEffect(() => {
        if (socket.current != null) {
            return;
        }

        window.onload = function () {
            socket.current = new WebSocket(path)
            socket.current.onopen = function (_) {
                console.info("ws opened...");

                // starting ping/pong messaging
                socket.current!.send(JSON.stringify({
                    message_type: PingMessage
                }));
            }
            socket.current.onclose = function (_) {
                console.info("ws closed...");
                clearTimeout(pingTimeout.current)
                socket.current = null;
            }
            socket.current.onmessage = function (e) {
                const message = JSON.parse(e.data) as WSMessage;
                // console.log("ws got message", message.message_type, message.content)
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
                if (onError) onError()
            }
        }

        return () => {
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
