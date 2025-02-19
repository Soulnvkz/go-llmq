import style from "./chat.module.scss"

import MessagesContainer from "../../containers/messagesContainer";
import Button from "../button/button";

interface ChatProps {
    chatName: string
    messages: any
    current: string

    inputRef: React.Ref<HTMLInputElement>
    currentRef: React.Ref<HTMLDivElement>

    onCancel: () => void
    onSend: () => void
}

function Chat({
    chatName,
    messages,
    current,
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
            <MessagesContainer ref={currentRef} messages={messages} current={current} />
            <div className={style.inputContainer}>
                <input ref={inputRef} type="text" className={style.messageInput} placeholder="Type your message..." />
                <Button text="Cancel" onClick={onCancel} />
                <Button text="Send" onClick={onSend} />
            </div>
        </div>)
}

export default Chat