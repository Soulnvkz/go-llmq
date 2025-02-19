import style from "./messages.module.scss"

interface MessagesProps  {
}

function Messages(props: React.PropsWithChildren<MessagesProps>) {
    return (
       <div className={style.container}>
            {props.children}
       </div>
    )
}

export default Messages