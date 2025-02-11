import style from "./messagesContainer.module.scss"

interface MessagesContainerProps  {
}

function MessagesContainer(props: React.PropsWithChildren<MessagesContainerProps>) {
    return (
       <div className={style.container}>
            {props.children}
       </div>
    )
}

export default MessagesContainer