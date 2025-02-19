import { useCallback, useContext, useEffect } from "react";

import { CancelMessage, CompletitionsEnd, CompletitionsMessage, CompletitionsNext, CompletitionsQueue, CompletitionsStart, useWebSocket, WSMessage } from "./useWebSocket";
import { WSContext } from "../state/WSContext";

interface Props {
    onQueue: () => void;
    onStart: () => void;
    onNext: (next: string) => void;
    onEnd: () => void;
}

export function useCompletions({
    onQueue,
    onStart,
    onNext,
    onEnd
}: Props) {
    const { send, addOnMessageCallback, removeOnMessageCallback } = useContext(WSContext)

    const onMessage = useCallback(function onMessage(message: WSMessage) {
            console.log("!!!")
            switch (message.message_type) {
                case CompletitionsQueue:
                    onQueue()
                    break
                case CompletitionsStart:
                    onStart()
                    break
                case CompletitionsNext:
                    onNext(message.content!)
                    break
                case CompletitionsEnd:
                    onEnd()
                    break
                default:
                    console.info("unsupported message type", message.message_type)
                    break
            }
        }
    , [])

    useEffect(() => {
        addOnMessageCallback(onMessage)

        return () => {
            removeOnMessageCallback(onMessage)
        }
    }, [onMessage])

    function request(message: string) {
        send(JSON.stringify({
            message_type: CompletitionsMessage,
            content: message
        }))
    }

    function cancel() {
        send(JSON.stringify({
            message_type: CancelMessage,
        }))
    }

    return {
        request,
        cancel
    }
}