import { useRef, useState } from "react"

import Chat from "../components/chat/chat"
import { useCompletions } from "../hooks/useCompletions"
import Message from "../models/Message"

function ChatContainer() {
    const index = useRef(0)
    const [messages, setMessages] = useState<Message[]>([])
    const [current, setCurrent] = useState("")
    const [isQueue, setQueue] = useState(false)

    const currentRef = useRef<HTMLDivElement>(null)
    const inputRef = useRef<HTMLInputElement>(null)

    const { cancel, request } = useCompletions({
        onQueue() {
            setQueue(true)
        },
        onStart() {
            setCurrent("")
        },
        onNext(next: string) {
            setQueue(false)
            setCurrent(prev => prev + next)
        },
        onEnd() {
            const text = currentRef.current!.innerHTML.replace("<span>", "").replace("</span>", "")
            index.current = index.current + 1

            setMessages(prev => [...prev, {
                id: index.current,
                isUser: false,
                text: text
            }])

            setCurrent("")
        },
    })

    const isCanCancel = current.length > 0 || isQueue

    function onCancel(){
        cancel()

        setCurrent("")
        setQueue(false)
    }

    function onSend() {
        const message = inputRef.current?.value ?? ""
        if (message.length === 0) {
            return;
        }

        index.current = index.current + 1;
        inputRef.current!.value = ""
        setMessages(prev => [...prev, {
            id: index.current,
            isUser: true,
            text: message
        }]);
        request(message)
    }

    return (<Chat
        chatName="Text Chat #1"
        messages={messages}
        current={current}
        inputRef={inputRef}
        currentRef={currentRef}
        isQueue={isQueue}
        isCanCancel={isCanCancel}
        onCancel={onCancel}
        onSend={onSend}
    />)
}

export default ChatContainer