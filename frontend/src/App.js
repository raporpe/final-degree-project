import './App.css';
import FloorSelector from './FloorSelector';
import React from 'react';
import SidePanel from './SidePanel';
import FloorMap from './FloorMap';

let fixedFloors = new Map();
fixedFloors.set(3, "Salas estudio")
fixedFloors.set(2, "Hemeroteca")
fixedFloors.set(1, "Biblioteca")
fixedFloors.set(0, "Planta baja")
fixedFloors.set(-1, "Sótano")

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
    // this.setState({currentRoom: null}) TODO: implementar cuando tengamos el selector de room
  }

  handleRoomChange = (newRoom) => {
    this.setState({currentRoom: newRoom}) 
  }

  render() {
    return (
      <div>
        <SidePanel currentRoom={this.state.currentRoom} onRoomChange={this.handleRoomChange}></SidePanel>
        <FloorSelector currentFloor={this.state.currentFloor} floors={fixedFloors} onFloorChange={this.handleFloorChange}></FloorSelector>
        <FloorMap currentFloor={this.state.currentFloor}></FloorMap>
      </div>
    );
  }

}

export default App;
