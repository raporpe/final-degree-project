import React from 'react';
import './Chart.css'
import { Bar } from "@visx/shape";
import { Group } from "@visx/group";
import { scaleBand, scaleLinear } from "@visx/scale";


class Chart extends React.Component {


    render() {
        let width = this.props.width;
        let height = this.props.height;
        let verticalMargin = 20;

        let xMax = width;
        let yMax = height - verticalMargin;

        let xScale = scaleBand({
            range: [0, xMax],
            round: true,
            domain: Object.entries(this.props.data).map((d) => d[0]),
            padding: 0.1
        });


        let yScale = scaleLinear({
            range: [yMax, 0],
            round: true,
            domain: [0, Math.max(...Object.entries(this.props.data).map((d) => d[1] * 100))]
        });

        return (
            <svg width={width} height={height}>
                <rect width={width} height={height} fill="url(#blue)" rx={14} />
                <Group top={verticalMargin / 2}>
                    {Object.entries(this.props.data).map((d) => {
                        const date = d[0]
                        const value = d[1]
                        const barWidth = xScale.bandwidth();
                        const barHeight = yMax - (yScale(value * 100) ?? 0);
                        const barX = xScale(date);
                        const barY = yMax - barHeight;
                        const color = "rgba(80, 171, 255, 1)"
                        return (
                            <Bar
                                key={`bar-${date}`}
                                x={barX}
                                y={barY}
                                rx={6}
                                width={barWidth}
                                height={barHeight}
                                fill={color}
                            />
                        );
                    })}
                </Group>
            </svg>
        );
    }
}


export default Chart;