import React from "react";
import "./FloorMap.css";


class FloorMap extends React.Component {
    render () {
        return(
            <div className="map-main">Mostrando mapa de la floor {this.props.currentFloor}</div>
        )
    }
}

export default FloorMap;