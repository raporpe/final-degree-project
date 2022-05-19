import React, { useMemo } from 'react';
import './Chart.css'
import { Bar } from "@visx/shape";
import { Group } from "@visx/group";
import { scaleBand, scaleLinear } from "@visx/scale";
import letterFrequency, {
    LetterFrequency
} from "@visx/mock-data/lib/mocks/letterFrequency";

const data = letterFrequency.slice(10);


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
            domain: data.map((d) => d.letter),
            padding: 0.1
        });


        let yScale = scaleLinear({
            range: [yMax, 0],
            round: true,
            domain: [0, Math.max(...data.map((d) => d.frequency * 100))]
        });

        return (
            <svg width={width} height={height}>
                <rect width={width} height={height} fill="url(#blue)" rx={14} />
                <Group top={verticalMargin / 2}>
                    {data.map((d) => {
                        const barWidth = xScale.bandwidth();
                        const barHeight = yMax - (yScale(d.frequency * 100) ?? 0);
                        const barX = xScale(d.letter);
                        const barY = yMax - barHeight;
                        const color = "rgba(80, 171, 255, 1)"
                        console.log("ym" + yMax);
                        return (
                            <Bar
                                key={`bar-${d.letter}`}
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