import './App.css'
import ChatContainer from './containers/chatContainer'
import { WSProvider } from './state/WSContext'

function App() {
  return (
    <WSProvider>
      <ChatContainer />
    </WSProvider>
  )
}

export default App