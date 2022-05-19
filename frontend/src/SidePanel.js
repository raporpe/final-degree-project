import React from "react";
import "./SidePanel.css";
import Chart from './Chart';


class SidePanel extends React.Component {
    //constructor(props) {
    //    super(props)
    //}

    render() {
        if (this.props.currentRoom === null) {
            return (
                <div className="sidebar">
                    <img className="sidebar-image" alt="biblioteca uc3m" src="/placeholder.png"></img>
                    <div className="sidebar-content">
                        <div className="sidebar-nodata">Selecciona una sala en el mapa</div>
                    </div>
                </div>
            )
        } else {
            return (
                <div className="sidebar">
                    <img className="sidebar-image" alt="biblioteca uc3m" src="/room.png"></img>
                    <div className="sidebar-content">
                        <div className="sidebar-title">{this.props.currentRoom}</div>
                        <div className="sidebar-ocupacion">70% de ocupacion</div>
                        <div className="sidebar-chart">
                            <Chart room={this.props.currentRoom} width={325} height={125}></Chart>
                        </div>
                    </div>
                </div>
            )
        }

    }
}

export default SidePanel;
