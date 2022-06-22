import React from "react";
import "./FloorMap.css";
import { Flex, C1, C2, C3, C4, C5, C6, C11, C21, C31, C41, C51, C12, C22, C32, C42, C52, FloorContainer } from "./FloorMapStyled.js";

class FloorMap extends React.Component {

    handleColor(occupation) {

        if (occupation >= 90) {
            return "tomato";
        }
        if (occupation >= 75 && occupation < 90) {
            return "orange";
        }

        if (occupation >= 50 && occupation < 75) {
            return "gold";
        }
        if (occupation >= 0 && occupation < 50) {
            return "aquamarine";
        }
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
        //array con las ocupaciones de cada habitacion
        let occupationsFloor1 = [10, 51, 78, 90] //planta baja
        let occupationsFloor2 = [80, 57]  //planta 1
        let occupationsFloor3 = [21] //planta 2
        //this.rooms[this.props.currentFloor].forEach((r) => {
        //    outputRooms.push(<div onClick={() => this.props.onRoomChange(r)}>{r}</div>)
        //})


        let ret = <div>
                <FloorContainer>
                    <C1></C1>
                    <div style={{width:"65%", height:"100%"}}>
                        <Flex style={{width:"100%", height:"33.3%"}}>
                            <C2 onClick={() => this.props.onRoomChange("entrada")} inputColor={this.handleColor(occupationsFloor1[0])}>Entrada {occupationsFloor1[0]}%

                            </C2>
                            <C3 onClick={() => this.props.onRoomChange("wc")} inputColor={this.handleColor(occupationsFloor1[1])}>WC {occupationsFloor1[1]}%</C3>
                        </Flex>
                        <Flex style={{width:"100%", height:"33.3%"}}>
                            <C4></C4>
                            <C5 onClick={() => this.props.onRoomChange("makerspace")} inputColor={this.handleColor(occupationsFloor1[2])}>Makerspace {occupationsFloor1[2]}%</C5>
                        </Flex>
                        <Flex style={{width:"100%", height:"33.3%"}}>
                            <C6 onClick={() => this.props.onRoomChange("sala_de_trabajo")} inputColor={this.handleColor(occupationsFloor1[3])}>Sala de trabajo {occupationsFloor1[3]}%</C6>
                        </Flex>
                    </div>
                </FloorContainer>
             </div>
        
        if (this.props.currentFloor === 1) {
            ret = <div>
            <FloorContainer>
                <C11 inputColor={this.handleColor(occupationsFloor1[0])}>Sala silenciosa {occupationsFloor1[0]}%</C11>
                    <Flex style={{ width: "65%", height: "100%" }}>
                        <div style={{ width: "50%", height: "100%" }}>
                            <C21 onClick={() => this.props.onRoomChange("entrada")} inputColor={this.handleColor(occupationsFloor1[0])}></C21>
                            <C41></C41>
                            <C51 onClick={() => this.props.onRoomChange("makerspace")} inputColor={this.handleColor(occupationsFloor1[0])}></C51>
                        </div>
                        <C31 onClick={() => this.props.onRoomChange("wc")} inputColor={this.handleColor(occupationsFloor1[1])}>Sala de trabajo {occupationsFloor2[1]}%</C31>
                    </Flex>
            </FloorContainer>
</div>
        }
        if (this.props.currentFloor === 2) {
            ret = <div>
            <FloorContainer>
                <C12 inputColor={this.handleColor(occupationsFloor3[0])}></C12>
                <Flex style={{ width: "65%", height: "100%" }}>
                    <div style={{ width: "50%", height: "100%" }}>
                        <C22 onClick={() => this.props.onRoomChange("entrada")} inputColor={this.handleColor(occupationsFloor3[0])}>Sala de trabajo {occupationsFloor3[0]}%</C22>
                        <C42></C42>
                        <C52 onClick={() => this.props.onRoomChange("makerspace")} inputColor={this.handleColor(occupationsFloor3[0])}></C52>
                    </div>
                    <C32 onClick={() => this.props.onRoomChange("wc")} inputColor={this.handleColor(occupationsFloor3[0])}></C32>
                </Flex>
            </FloorContainer>
        </div>
        }

        return (
            <>
                {ret}
            </>
        )
    }
}

export default FloorMap;