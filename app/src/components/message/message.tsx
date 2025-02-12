import React from "react";
import style from "./message.module.scss"

interface MessageProps {
    isUser: boolean;
    text?: string;
    last?: boolean;
}

const Message = React.forwardRef(function({text, isUser, last = false}: MessageProps, ref: React.Ref<HTMLDivElement>) {
    let messageStyle = `${style.message}`
    if(isUser){
        messageStyle+=` ${style.user}`;
    }else{
        messageStyle+=` ${style.bot}`;
    }

    let stylex = {}
    if(last) {
        stylex = {
            display:"none"
        }
    }

    return (
       <div ref={ref} style={stylex} className={messageStyle}>
            <span>{text}</span>
       </div>
    )
})



export default Message