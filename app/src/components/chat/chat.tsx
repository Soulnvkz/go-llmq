import style from "./chat.module.scss"

import MessagesContainer from "../../containers/messagesContainer";
import Button from "../button/button";
import CancelButton from "../button/cancelButton";

interface ChatProps {
    chatName: string
    messages: any
    current: string

    inputRef: React.Ref<HTMLInputElement>
    currentRef: React.Ref<HTMLDivElement>

    isQueue: boolean
    isCanCancel: boolean

    onCancel: () => void
    onSend: () => void
}

function Chat({
    chatName,
    messages,
    current,
    isQueue,
    isCanCancel,
    inputRef,
    currentRef,
    onCancel,
    onSend,
}: ChatProps) {
    return (
        <div className={style.chatContainer}>
            <div className={style.chatHeader}>
                <h1>{chatName}</h1>
            </div>
            <MessagesContainer ref={currentRef} messages={messages} current={current} isQueue={isQueue} />
            <div className={style.inputContainer}>
                <input ref={inputRef} type="text" className={style.messageInput} placeholder="Type your message..." />
                {isCanCancel ? <CancelButton onClick={onCancel} /> : <Button text="Send" onClick={onSend} />}
            </div>
        </div>)
}

export default Chat