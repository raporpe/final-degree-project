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
            "https://ey2qdxofkj.execute-api.eu-west-1.amazonaws.com/historic?" + new URLSearchParams({
                from_time: startTime.toISOString(),
                to_time: currentTime.toISOString()
            }))
            .then((res) => res.json())
            .then((json) => {
                console.log(json);

                Object.entries(json.rooms).forEach(room => {
                    let compressed = {}
                    let counter = {}

                    for (let i = 0; i < 24; i++) {
                        compressed[i] = 0.0;
                        counter[i] = 0.0;
                    }

                    Object.entries(room[1]).forEach((d) => {
                        let toDate = new Date(d[0])
                        compressed[toDate.getHours()] = compressed[toDate.getHours()] + d[1]*1.0;
                        counter[toDate.getHours()] += 1.0
                    })

                    // Divide to calculate average
                    Object.entries(compressed).forEach((e) => {
                        if (counter[e[0]] > 0) {
                            compressed[e[0]] = (compressed[e[0]]*1.0) / (counter[e[0]]*1.0)
                        }
                    })

                    console.log(compressed)

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

        const current = new Date();
        const currentHour = current.getHours(); 

        return (
            <div className="sidebar">
                <img className="sidebar-image" alt="biblioteca uc3m" src="/room.png"></img>
                <div className="sidebar-content">
                    <div className="sidebar-title">{roomNames[this.props.currentRoom]}</div>
                    
                    <div className="sidebar-ocupacion">{parseInt(this.state.data.rooms[this.props.currentRoom][currentHour])}% de ocupacion</div>
                    <div className="sidebar-chart">
                        <Chart data={this.state.data.rooms[this.props.currentRoom]} width={325} height={150}></Chart>
                    </div>
                </div>
            </div>
        )



    }
}

export default SidePanel;
