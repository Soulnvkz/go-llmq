import React from "react"
import { useMemo } from "react"

import IMessage from "../models/Message"
import Messages from "../components/message/messages"
import Message from "../components/message/message"

interface messagesContainerProps {
    messages: IMessage[]
    current: string
}

const MessagesContainer = React.forwardRef(function({messages, current}: messagesContainerProps, ref: React.Ref<HTMLDivElement>) {
    const MessagesCached = useMemo(() =>
        messages.map(m => <Message key={m.id} text={m.text} isUser={m.isUser} />), [messages])

    return (
        <Messages>
            {MessagesCached.map(x => x)}
            <Message ref={ref} text={current} isUser={false} />
        </Messages>
    )
})

export default MessagesContainer