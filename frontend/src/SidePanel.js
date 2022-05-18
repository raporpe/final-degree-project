import React from "react";
import "./SidePanel.css";
import Chart from './Chart';


class SidePanel extends React.Component {
    //constructor(props) {
    //    super(props)
    //}

    render() {
        return (
            <div className="sidebar">
                <img className="sidebar-image" alt="biblioteca uc3m" src="/biblioteca.webp"></img>
                <div className="sidebar-content">
                    <div className="sidebar-title">Sala de estudio</div>
                    <div className="sidebar-ocupacion">70% de ocupacion</div>
                    <div className="sidebar-chart">
                        <Chart room={this.props.currentRoom}></Chart>
                    </div>
                </div>
            </div>
        )
    }
}

export default SidePanel;
