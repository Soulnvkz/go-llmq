import React from "react"

import style from "./message.module.scss"

interface MessageProps {
    isUser: boolean
    text?: string
}

const Message = React.forwardRef(function({text, isUser}: MessageProps, ref: React.Ref<HTMLDivElement>) {
    const classes = `${style.message} ${isUser ? style.user: style.bot}`
    if(!text || text.length === 0) {
        return
    }

    return (
       <div ref={ref} className={classes}>
            <span>{text}</span>
       </div>
    )
})



export default Message