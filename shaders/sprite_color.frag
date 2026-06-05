#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 1) in vec4 fragColor;
layout(location = 0) out vec4 outColor;
layout(location = 1) out vec4 outEmissive;

layout(binding = 0) uniform sampler2D albedoTexture;
layout(push_constant) uniform MaterialPass {
    float emissive;
} materialPass;

void main() {
    vec4 albedo = texture(albedoTexture, fragUV) * fragColor;
    outColor = albedo;
    outEmissive = vec4(albedo.rgb * max(materialPass.emissive, 0.0), albedo.a);
}
