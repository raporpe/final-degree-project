import './App.css';
import FloorSelector from './FloorSelector';
import React from 'react';
import SidePanel from './SidePanel';

const floors = new Map()
floors.set(3, "Salas estudio")
floors.set(2, "Hemeroteca")
floors.set(1, "Biblioteca")
floors.set(0, "Planta baja")
floors.set(-1, "Sótano")

class App extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      currentFloor: 0,
      currentRoom: null,
    }
  }

  handleFloorChange = (newFloor) => {
    this.setState({currentFloor: newFloor}) 
  }

  handleRoomChange = (newRoom) => {
    this.setState({currentRoom: newRoom}) 
  }

  render() {
    return (
      <div>
        <SidePanel currentRoom={this.state.currentRoom} onRoomChange={this.handleRoomChange}></SidePanel>
        <FloorSelector currentFloor={this.state.currentFloor} floors={floors} onFloorChange={this.handleFloorChange}></FloorSelector>
      </div>
    );
  }

}

export default App;
