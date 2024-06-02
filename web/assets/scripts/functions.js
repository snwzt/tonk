// draw tank
function draw(dir) {
    context.clearRect(0, 0, canvas.width, canvas.height);
    if (dir === "up") {
        context.drawImage(svgImageTankUp, xPosition, yPosition);
        context.drawImage(svgImageTankTop, xPosition + (svgImageTankBottom.width / 5.5), yPosition + (svgImageTankBottom.height / 3.5));
    } else if (dir === "down") {
        context.drawImage(svgImageTankDown, xPosition, yPosition);
        context.drawImage(svgImageTankTop, xPosition + (svgImageTankBottom.width / 5.5), yPosition + (svgImageTankBottom.height / 3.5));
    } else{
        context.drawImage(svgImageTankBottom, xPosition, yPosition);
        context.drawImage(svgImageTankTop, xPosition + (svgImageTankBottom.width / 5.5), yPosition + (svgImageTankBottom.height / 4.5));
    }
}

// for angle in 0 45 90 135 180 225 270 315; do
//   mv "tank-up-${angle}.svg" "tank-up-pu-pu-${angle}.svg"
//   mv "tank-up-${angle} (1).svg" "tank-up-pu-pi-${angle}.svg"
//   mv "tank-up-${angle} (2).svg" "tank-up-pu-or-${angle}.svg"
//   mv "tank-up-${angle} (3).svg" "tank-up-pu-br-${angle}.svg"
//   mv "tank-up-${angle} (4).svg" "tank-up-pu-bl-${angle}.svg"
// done

