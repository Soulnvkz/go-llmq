import style from "./button.module.scss"

interface ButtonProps extends React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> {
}

function CancelButton(props: ButtonProps) {
    return (
        <button className={`${style.button} ${style.cancel}`} {...props}>
            <div className={style.cancelBox}></div>
        </button>
    )
}

export default CancelButton