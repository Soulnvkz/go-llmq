import { useEffect, useRef } from "react";

interface Params {
    path: string;
    onMessage: (message: string) => void; 
}

function useWebSocket({
    path,
    onMessage,
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
                socket.current!.send("ping");
            }
            socket.current.onclose = function (_) {
                console.info("ws closed...");
                clearTimeout(pingTimeout.current)
                socket.current = null;
            }
            socket.current.onmessage = function (e) {
                console.info("ws got message...", e.data);

                if (e.data === "pong") {
                    pingTimeout.current = setTimeout(() => {
                        socket.current!.send("ping");
                    }, 1000);
                    return;
                }

                onMessage(e.data);
            }
            socket.current.onerror = function (_) {
                console.error("ws error...");
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

export default useWebSocket;