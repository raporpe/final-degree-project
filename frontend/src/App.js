import './App.css';
import FloorSelector from './FloorSelector';
import React from 'react';

const pre_floors = new Map()
pre_floors.set(3, "Salas estudio")
pre_floors.set(2, "Hemeroteca")
pre_floors.set(1, "Biblioteca")
pre_floors.set(0, "Planta baja")
pre_floors.set(-1, "Sótano")

function App() {
  return (
    <FloorSelector floors={pre_floors}></FloorSelector>
  );
}

export default App;
