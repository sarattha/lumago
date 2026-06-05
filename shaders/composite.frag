#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 0) out vec4 outColor;

layout(binding = 0) uniform sampler2D sceneColor;
layout(binding = 1) uniform sampler2D sceneNormal;
layout(binding = 2) uniform sampler2D lightBuffer;
layout(binding = 3) uniform CompositeUniforms {
    uint debugView;
} composite;

void main() {
    vec4 color = texture(sceneColor, fragUV);
    vec4 normal = texture(sceneNormal, fragUV);
    vec4 light = texture(lightBuffer, fragUV);

    if (composite.debugView == 1) {
        outColor = color;
    } else if (composite.debugView == 2) {
        outColor = normal;
    } else if (composite.debugView == 3) {
        outColor = light;
    } else {
        outColor = vec4(color.rgb * light.rgb, color.a);
    }
}
