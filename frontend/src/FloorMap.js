import React from "react";
import "./FloorMap.css";
import {Section} from "./FloorMapStyled.js";

class FloorMap extends React.Component {

    handleColor(){
        return;
    }

    rooms = {
        3: ["Sala A", "Sala B"],
        2: ["Estudio A", "Estudio B"],
        1: ["Biblioteca A", "Biblioteca B", "Estudio 3"],
        0: ["Entrada", "Baños", "Makerspace", "Colaboración"],
        [-1]: ["Clases A", "Clases B", "Sala de presentaciones"],
    }

    render() {
        let outputRooms = []
        this.rooms[this.props.currentFloor].forEach((r) => {
            outputRooms.push(<div onClick={() => this.props.onRoomChange(r)}>{r}</div>)
        })

        return (
            <div>
                <div className="map-main">Mostrando mapa de la floor {this.props.currentFloor}
                    <div>
                        {outputRooms}
                    </div>

                </div>
            </div>
        )
    }
}

export default FloorMap;