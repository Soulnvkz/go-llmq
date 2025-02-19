import React from "react"
import { useMemo } from "react"

import IMessage from "../models/Message"
import Messages from "../components/message/messages"
import Message from "../components/message/message"
import Queue from "../components/message/queue"

interface messagesContainerProps {
    messages: IMessage[]
    current: string
    isQueue: boolean
}

const MessagesContainer = React.forwardRef(function({messages, current, isQueue}: messagesContainerProps, ref: React.Ref<HTMLDivElement>) {
    const ContextMessages = useMemo(() =>
        messages.map(m => <Message key={m.id} text={m.text} isUser={m.isUser} />), [messages])

    return (
        <Messages>
            {ContextMessages.map(x => x)}
            <Message ref={ref} text={current} isUser={false} />
            {isQueue && <Queue />}
        </Messages>
    )
})

export default MessagesContainer