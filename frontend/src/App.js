import './App.css';
import FloorSelector from './FloorSelector';
import React from 'react';
import SidePanel from './SidePanel';
import FloorMap from './FloorMap';
import ZoneSelector from './ZoneSelector';
import Chart from './Chart'
import './App.css'

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
      currentZone: "Leganés",
    }
  }

  handleFloorChange = (newFloor) => {
    this.setState({currentFloor: newFloor})
    this.setState({currentRoom: null})
  }

  handleRoomChange = (newRoom) => {
    this.setState({currentRoom: newRoom}) 
  }

  render() {
    return (
      <div>
        <div className="section">
          <SidePanel currentRoom={this.state.currentRoom} viewBox="0 0 400 325"></SidePanel>
          <FloorMap currentFloor={this.state.currentFloor} onRoomChange={this.handleRoomChange}></FloorMap>
        </div>
        <FloorSelector currentFloor={this.state.currentFloor} floors={fixedFloors} onFloorChange={this.handleFloorChange}></FloorSelector>
        <ZoneSelector currentZone={this.state.currentZone}></ZoneSelector>        
      </div>
    );
  }

}

export default App;
