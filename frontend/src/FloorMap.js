import React from "react";
import "./FloorMap.css";
import { Flex, C1, C2, C3, C4, C5, C6 } from "./FloorMapStyled.js";

class FloorMap extends React.Component {

    handleColor() {

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
        //this.rooms[this.props.currentFloor].forEach((r) => {
        //    outputRooms.push(<div onClick={() => this.props.onRoomChange(r)}>{r}</div>)
        //})

        return (
            <div>
                <div>

                    {outputRooms}
                    <div>
                        <Flex>
                            <C1></C1>
                            <div>
                                <Flex>
                                    <C2 onClick={() => this.props.onRoomChange("entrada")} inputColor={this.handleColor(/*aquí numerito de la hab */)}>Entrada</C2>
                                    <C3 onClick={() => this.props.onRoomChange("wc")} inputColor={this.handleColor()}>WC</C3>
                                </Flex>
                                <Flex>
                                    <C4></C4>
                                    <C5 onClick={() => this.props.onRoomChange("makerspace")} inputColor={this.handleColor()}>Makerspace</C5>
                                </Flex>
                                <Flex>
                                    <C6 onClick={() => this.props.onRoomChange("sala_de_trabajo")} inputColor={this.handleColor()}>Sala de trabajo</C6>
                                </Flex>
                            </div>
                        </Flex>
                    </div>

                </div>
            </div >
        )
    }
}

export default FloorMap;