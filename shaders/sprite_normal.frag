#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 0) out vec4 outNormal;

layout(binding = 0) uniform sampler2D normalTexture;
layout(push_constant) uniform NormalPass {
    uint hasNormalMap;
} normalPass;

void main() {
    if (normalPass.hasNormalMap == 0) {
        outNormal = vec4(0.5, 0.5, 1.0, 1.0);
        return;
    }

    vec3 normal = texture(normalTexture, fragUV).xyz;
    outNormal = vec4(normalize(normal * 2.0 - 1.0) * 0.5 + 0.5, 1.0);
}
