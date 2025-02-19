import { useRef, useState } from "react"

import Chat from "../components/chat/chat"
import { useCompletions } from "../hooks/useCompletions"
import Message from "../models/Message"

function ChatContainer() {
    const index = useRef(0)
    const [messages, setMessages] = useState<Message[]>([])
    const [current, setCurrent] = useState("")

    const currentRef = useRef<HTMLDivElement>(null)
    const inputRef = useRef<HTMLInputElement>(null)

    const { cancel, request } = useCompletions({
        onQueue() {
            setCurrent("queue...")
        },
        onStart() {
            setCurrent("")
        },
        onNext(next: string) {
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

    return (<Chat
        chatName="Text Chat #1"
        messages={messages}
        current={current}
        inputRef={inputRef}
        currentRef={currentRef}
        onCancel={() => { cancel() }}
        onSend={() => {
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
        }}
    />)
}

export default ChatContainer