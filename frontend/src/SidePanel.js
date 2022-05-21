import React from "react";
import "./SidePanel.css";
import Chart from './Chart';
import ContentLoader from "react-content-loader"

const roomNames = {
    sala_de_trabajo: "Sala de trabajo",
    wc: "Baños",
    makerspace: "MakerSpace",
    entrada: "Entrada",
}

class SidePanel extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            loaded: false,
            data: null,
        }
    }

    componentDidMount() {
        let startTime = new Date();
        startTime.setHours(0, 0, 0, 0);
        console.log(startTime.toISOString())
        console.log(startTime)
        let currentTime = new Date();

        fetch(
            "https://tfg-api.raporpe.dev/v1/historic?" + new URLSearchParams({
                from_time: startTime.toISOString(),
                to_time: currentTime.toISOString()
            }))
            .then((res) => res.json())
            .then((json) => {
                console.log(json);

                Object.entries(json.rooms).forEach(room => {
                    let compressed = {}

                    for (let i = 0; i < 24; i++) {
                        compressed[i] = 0;
                    }

                    Object.entries(room[1]).forEach((d) => {
                        let toDate = new Date(d[0])
                        compressed[toDate.getHours()] = compressed[toDate.getHours()] + (d[1]*1.0 / 60);

                    })
                    json.rooms[room[0]] = compressed;
                })
                this.setState({
                    data: json,
                    loaded: true
                });
                console.log(this.state.data)
            })
    }

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
        }

        if (!this.state.loaded) {
            return (
                <div className="sidebar">
                    <ContentLoader
                        speed={2}
                        width={200}
                        height={400}
                        viewBox="0 0 400 400"
                        backgroundColor="#f3f3f3"
                        foregroundColor="#ecebeb"
                        {...this.props}
                    >
                        <rect x="3" y="1" rx="0" ry="0" width="196" height="124" />
                        <rect x="2" y="144" rx="0" ry="0" width="161" height="21" />
                        <rect x="3" y="179" rx="0" ry="0" width="128" height="17" />
                    </ContentLoader>
                </div>
            )
        }

        return (
            <div className="sidebar">
                <img className="sidebar-image" alt="biblioteca uc3m" src="/room.png"></img>
                <div className="sidebar-content">
                    <div className="sidebar-title">{roomNames[this.props.currentRoom]}</div>
                    <div className="sidebar-ocupacion">{70}% de ocupacion</div>
                    <div className="sidebar-chart">
                        <Chart data={this.state.data.rooms[this.props.currentRoom]} width={325} height={125}></Chart>
                    </div>
                </div>
            </div>
        )



    }
}

export default SidePanel;
