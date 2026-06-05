#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 0) out vec4 outColor;

void main() {
    vec2 tile = floor(fragUV * 8.0);
    float checker = mod(tile.x + tile.y, 2.0);
    vec3 colorA = vec3(0.95, 0.82, 0.32);
    vec3 colorB = vec3(0.16, 0.56, 0.70);
    outColor = vec4(mix(colorA, colorB, checker), 1.0);
}
