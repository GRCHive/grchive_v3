# Copyright 2017 The Bazel Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

def _impl(ctx):
    """Quick and non-hermetic rule to build a Jekyll site."""
    source = ctx.actions.declare_directory(ctx.attr.name + "-srcs")
    output = ctx.actions.declare_directory(ctx.attr.name + "-build")

    ctx.actions.run_shell(
        inputs = ctx.files.srcs,
        outputs = [source],
        command = ("mkdir -p %s\n" % (source.path)) +
                  "\n".join([
                      "tar xf %s -C %s" % (src.path, source.path)
                      for src in ctx.files.srcs
                  ]),
    )

    ctx.actions.run_shell(
        inputs = [source],
        outputs = [output],
        use_default_shell_env = True,
        command = "JEKYLL_ENV={0} jekyll build -s {1} -d {2}".format(
            ctx.attr.env,
            source.path,
            output.path,
        )
    )

    ctx.actions.run(
        inputs = [output],
        outputs = [ctx.outputs.out],
        executable = "tar",
        arguments = ["cfh", ctx.outputs.out.path, "-C", output.path, ".", "--transform", "s,^,www/,"],
    )

    ctx.actions.expand_template(
        template = ctx.file._jekyll_serve_tpl,
        output = ctx.outputs.executable,
        substitutions = {
            "%{workspace_name}": ctx.workspace_name,
            "%{source_dir}": source.short_path,
        },
        is_executable = True,
    )
    return [DefaultInfo(files=depset([ctx.outputs.out]), runfiles = ctx.runfiles(files = [source, output]))]

jekyll_build = rule(
    implementation = _impl,
    executable = True,
    attrs = {
        "srcs": attr.label_list(allow_empty = False),
        "env": attr.string(default='development'),
        "_jekyll_serve_tpl": attr.label(
            default = ":jekyll_serve.sh.tpl",
            allow_single_file = True,
        ),
    },
    outputs = {"out": "%{name}.tar"},
)
