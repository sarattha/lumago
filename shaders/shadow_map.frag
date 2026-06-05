#version 450

layout(location = 0) out vec4 outShadow;

layout(push_constant) uniform ShadowMapPush {
    vec2 lightPosition;
    float lightRadius;
    uint segmentCount;
} shadowMap;

void main() {
    float normalizedDistance = 1.0;
    outShadow = vec4(normalizedDistance, normalizedDistance, normalizedDistance, 1.0);
}
