import { useEffect, useRef } from "react";
import "./InteractiveBG.css";

const CELL_SIZE = 45;
const EFFECT_RADIUS = CELL_SIZE;
const FADE_SPEED = 0.035;

export default function InteractiveBackground() {
  const canvasRef = useRef(null);
  const mouse = useRef({ x: -1000, y: -1000 });
  const animationFrame = useRef(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");

    let width = 0;
    let height = 0;
    let cols = 0;
    let rows = 0;
    let cells = [];

    const resize = () => {
      const dpr = window.devicePixelRatio || 1;

      width = window.innerWidth;
      height = window.innerHeight;

      canvas.width = width * dpr;
      canvas.height = height * dpr;
      canvas.style.width = `${width}px`;
      canvas.style.height = `${height}px`;

      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

      cols = Math.ceil(width / CELL_SIZE);
      rows = Math.ceil(height / CELL_SIZE);

      cells = Array.from({ length: cols * rows }, () => ({
        glow: 0,
      }));
    };

    const handleMouseMove = (event) => {
      mouse.current.x = event.clientX;
      mouse.current.y = event.clientY;
    };

    const handleMouseLeave = () => {
      mouse.current.x = -1000;
      mouse.current.y = -1000;
    };

    const draw = () => {
      ctx.clearRect(0, 0, width, height);

      for (let row = 0; row < rows; row++) {
        for (let col = 0; col < cols; col++) {
          const index = row * cols + col;

          const x = col * CELL_SIZE;
          const y = row * CELL_SIZE;

          const centerX = x + CELL_SIZE / 2;
          const centerY = y + CELL_SIZE / 2;

          const dx = centerX - mouse.current.x;
          const dy = centerY - mouse.current.y;

          const distance = Math.sqrt(dx * dx + dy * dy);

          if (distance < EFFECT_RADIUS) {
            const strength = 1 - distance / EFFECT_RADIUS;

            cells[index].glow = Math.max(
              cells[index].glow,
              strength
            );
          }

          cells[index].glow -= FADE_SPEED;

          if (cells[index].glow < 0) {
            cells[index].glow = 0;
          }

          const glow = cells[index].glow;

          
          ctx.strokeStyle = `rgba(255, 255, 255, ${
            0.035 + glow * 0.2
          })`;

          ctx.lineWidth = 1;

        

          // Свечение внутри клетки
          if (glow > 0.01) {
            const gradient = ctx.createRadialGradient(
              centerX,
              centerY,
              0,
              centerX,
              centerY,
              CELL_SIZE
            );

            gradient.addColorStop(
              0,
              `rgba(100, 180, 255, ${glow * 0.18})`
            );

            gradient.addColorStop(
              1,
              "rgba(100, 180, 255, 0)"
            );

            ctx.fillStyle = gradient;

            ctx.fillRect(
              x,
              y,
              CELL_SIZE,
              CELL_SIZE
            );
          }
        }
      }

      animationFrame.current = requestAnimationFrame(draw);
    };

    resize();
    draw();

    window.addEventListener("resize", resize);
    window.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseleave", handleMouseLeave);

    return () => {
      window.removeEventListener("resize", resize);
      window.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseleave", handleMouseLeave);

      cancelAnimationFrame(animationFrame.current);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="interactive-background"
    />
  );
}