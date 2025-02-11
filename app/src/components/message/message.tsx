import style from "./message.module.scss"

interface MessageProps {
    text: string;
    isUser: boolean
}

function Message({text, isUser}: MessageProps) {
    let messageStyle = `${style.message}`
    if(isUser){
        messageStyle+=` ${style.user}`;
    }else{
        messageStyle+=` ${style.bot}`;
    }

    return (
       <div className={messageStyle}>
            {text}
       </div>
    )
}

export default Message