import style from "./queue.module.scss"

function Queue() {
    return (<div className={style.thinkingIndicator}>
        <span className={style.dotFlashing}></span>
    </div>)
}

export default  Queue