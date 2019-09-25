// Copyright 2019 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package constants

const (
	EmailNotifyTemplate = `
<!doctype html>
<html>
<head>
	<meta name="viewport" content="width=device-width" />
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	<title>Simple Transactional Email</title>
	<style>
		/* -------------------------------------
            GLOBAL RESETS
        ------------------------------------- */
		/*All the styling goes here*/
		img {
			border: none;
			-ms-interpolation-mode: bicubic;
			max-width: 100%;
		}

		body {
			background-color: #eff0f5;
			font-family: Roboto, PingFang SC, Lantinghei SC, Helvetica Neue, Helvetica, Arial, Microsoft YaHei, 微软雅黑, STHeitiSC-Light, simsun, 宋体, WenQuanYi Zen Hei, WenQuanYi Micro Hei, sans-serif;
			-webkit-font-smoothing: antialiased;
			font-size: 14px;
			line-height: 1.4;
			margin: 0;
			padding: 0;
			-ms-text-size-adjust: 100%;
			-webkit-text-size-adjust: 100%;
			color: #576075;
		}

		table {
			border-collapse: separate;
			mso-table-lspace: 0pt;
			mso-table-rspace: 0pt;
			width: 100%; }
		table td {
			font-size: 14px;
			vertical-align: top;
		}

		/* -------------------------------------
            BODY & CONTAINER
        ------------------------------------- */

		.body {
			background-color: #eff0f5;
			width: 100%;
		}

		/* Set a max-width, and make it display as block so it will automatically stretch to that width, but will also shrink down on a phone or something */
		.container {
			display: block;
			margin: 0 auto !important;
			/* makes it centered */
			max-width: 780px;
			padding: 10px;
			padding-top: 80px;
			width: 780px;
		}

		/* This should also be a block element, so that it will fill 100% of the .container */
		.content {
			box-sizing: border-box;
			display: block;
			margin: 0 auto;
			max-width: 780px;
			padding: 10px;
		}

		/* -------------------------------------
            HEADER, FOOTER, MAIN
        ------------------------------------- */
		.main {
			background: #ffffff;
			border-radius: 2px;
			box-shadow: 0 1px 4px 0 rgba(73, 33, 173, 0.06), 0 4px 8px 0 rgba(35, 35, 36, 0.04);
			width: 100%;
		}

		.wrapper {
			box-sizing: border-box;
			padding: 48px;
		}

		.content-block {
			padding-bottom: 10px;
			padding-top: 10px;
		}

		.footer {
			clear: both;
			margin-top: 14px;
			text-align: center;
			width: 100%;
		}
		.footer td,
		.footer p,
		.footer span,
		.footer a {
			color: #8c96ad;
			font-size: 12px;
			text-align: center;
		}
		.gray {
			color: #8c96ad;
		}

		/* -------------------------------------
            TYPOGRAPHY
        ------------------------------------- */
		h1,
		h2,
		h3,
		h4 {
			color: #000000;
			font-weight: 400;
			line-height: 1.4;
			margin: 0;
			margin-bottom: 30px;
		}

		h1 {
			font-size: 35px;
			font-weight: 300;
			text-align: center;
			text-transform: capitalize;
		}

		p,
		ul,
		ol {
			font-size: 14px;
			font-weight: normal;
			line-height: 2;
			margin: 0;
			margin-bottom: 15px;
		}
		p li,
		ul li,
		ol li {
			list-style-position: inside;
			margin-left: 5px;
		}

		a {
			color: #8454cd;
			text-decoration: none;
		}

		/* -------------------------------------
            BUTTONS
        ------------------------------------- */
		.btn {
			box-sizing: border-box;
			width: 100%; }
		.btn > tbody > tr > td {
			padding-bottom: 15px; }
		.btn table {
			width: auto;
		}
		.btn table td {
			background-color: #ffffff;
			border-radius: 5px;
			text-align: center;
		}
		.btn a {
			background-color: #ffffff;
			border: solid 1px #3498db;
			border-radius: 5px;
			box-sizing: border-box;
			color: #3498db;
			cursor: pointer;
			display: inline-block;
			font-size: 14px;
			font-weight: bold;
			margin: 0;
			padding: 12px 25px;
			text-decoration: none;
			text-transform: capitalize;
		}

		.btn-primary table td {
			background-color: #3498db;
		}

		.btn-primary a {
			background-color: #3498db;
			border-color: #3498db;
			color: #ffffff;
		}

		/* -------------------------------------
            OTHER STYLES THAT MIGHT BE USEFUL
        ------------------------------------- */
		.last {
			margin-bottom: 0;
		}

		.first {
			margin-top: 0;
		}

		.align-center {
			text-align: center;
		}

		.align-right {
			text-align: right;
		}

		.align-left {
			text-align: left;
		}

		.clear {
			clear: both;
		}

		.mt0 {
			margin-top: 0;
		}

		.mb0 {
			margin-bottom: 0;
		}

		.preheader {
			color: transparent;
			display: none;
			height: 0;
			max-height: 0;
			max-width: 0;
			opacity: 0;
			overflow: hidden;
			mso-hide: all;
			visibility: hidden;
			width: 0;
		}

		.powered-by a {
			text-decoration: none;
		}

		hr {
			border: 0;
			border-bottom: 1px solid #eff0f5;
			margin: 50px 0 12px;
		}
		.linkBtn {
			border-radius: 2px;
			box-shadow: 0 1px 3px 0 rgba(73, 33, 173, 0.16), 0 1px 2px 0 rgba(52, 57, 69, 0.03);
			background-color: #8454cd;
			color: #fff;
			padding: 4px 20px;
		}
		.link {
			font-size: 12px;
			font-weight: normal;
			font-style: normal;
			font-stretch: normal;
			line-height: 28px;
			letter-spacing: normal;
		}
		.platform {
			font-size: 14px;
			font-weight: 500;
			font-style: normal;
			font-stretch: normal;
			line-height: 20px;
			letter-spacing: normal;
			color: #343945;
			margin-left: 12px;
		}
		.line1 {
			margin-top: 42px;
			margin-bottom: 16px;
		}
		.line2 {
			line-height: 2;
			margin-top: 16px;
			margin-bottom: 20px;
		}
		.line3 {
			margin-bottom: 40px;
			margin-top: 16px;
		}
		.line4 {
			margin-bottom: 0px;
		}
		.line5 {
			margin-top: 0px;
		}
		.line6 {
			margin-bottom: 0px;
		}

		/* -------------------------------------
            RESPONSIVE AND MOBILE FRIENDLY STYLES
        ------------------------------------- */
		@media only screen and (max-width: 620px) {
			table[class=body] h1 {
				font-size: 28px !important;
				margin-bottom: 10px !important;
			}
			table[class=body] p,
			table[class=body] ul,
			table[class=body] ol,
			table[class=body] td,
			table[class=body] span,
			table[class=body] a {
				font-size: 16px !important;
			}
			table[class=body] .wrapper,
			table[class=body] .article {
				padding: 10px !important;
			}
			table[class=body] .content {
				padding: 0 !important;
			}
			table[class=body] .container {
				padding: 0 !important;
				width: 100% !important;
			}
			table[class=body] .main {
				border-left-width: 0 !important;
				border-radius: 0 !important;
				border-right-width: 0 !important;
			}
			table[class=body] .btn table {
				width: 100% !important;
			}
			table[class=body] .btn a {
				width: 100% !important;
			}
			table[class=body] .img-responsive {
				height: auto !important;
				max-width: 100% !important;
				width: auto !important;
			}
		}

		/* -------------------------------------
            PRESERVE THESE STYLES IN THE HEAD
        ------------------------------------- */
		@media all {
			.ExternalClass {
				width: 100%;
			}
			.ExternalClass,
			.ExternalClass p,
			.ExternalClass span,
			.ExternalClass font,
			.ExternalClass td,
			.ExternalClass div {
				line-height: 100%;
			}
			.apple-link a {
				color: inherit !important;
				font-family: inherit !important;
				font-size: inherit !important;
				font-weight: inherit !important;
				line-height: inherit !important;
				text-decoration: none !important;
			}
			.btn-primary table td:hover {
				background-color: #34495e !important;
			}
			.btn-primary a:hover {
				background-color: #34495e !important;
				border-color: #34495e !important;
			}
		}

	</style>
</head>
<body class="">
<span class="preheader">This is preheader text. Some clients will show this text as a preview.</span>
<table role="presentation" border="0" cellpadding="0" cellspacing="0" class="body">
	<tr>
		<td>&nbsp;</td>
		<td class="container">
			<div class="content">
				<!-- START CENTERED WHITE CONTAINER -->
				<table role="presentation" class="main">

					<!-- START MAIN CONTENT AREA -->
					<tr>
						<td class="wrapper">
							<table role="presentation" border="0" cellpadding="0" cellspacing="0">
								<tr>
									<td>
										<p>
											<svg width="16px" height="16px" viewBox="0 0 16 16" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
												<!-- Generator: Sketch 52.2 (67145) - http://www.bohemiancoding.com/sketch -->
												<title>Rectangle</title>
												<desc>Created with Sketch.</desc>
												<g id="Navigation" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
													<g id="Admin-Navigation---我的工作台" transform="translate(-25.000000, -24.000000)">
														<rect id="Rectangle" fill="#FFFFFF" fill-rule="nonzero" x="0" y="0" width="64" height="64"></rect>
														<polygon id="Rectangle" fill="#FFFFFF" fill-rule="nonzero" points="0 0 64 0 64 900 0 900"></polygon>
														<rect id="Rectangle" fill-rule="nonzero" x="25" y="24" width="16" height="16"></rect>
														<g id="logo-new" transform="translate(25.000000, 24.060000)">
															<g id="Group">
																<path d="M3.24886005,7.29741333 L6.81243645,7.29741333 C7.14426097,7.29741333 7.43390666,7.23501333 7.68270831,7.11101333 C7.93150996,6.98701333 8.13519629,6.81794667 8.2937673,6.60408 C8.45260526,6.39021333 8.57006527,6.14568 8.64614732,5.86914667 C8.72196241,5.59341333 8.76013691,5.31048 8.76013691,5.02061333 C8.76013691,4.71741333 8.72196241,4.42754667 8.64614732,4.15154667 C8.57006527,3.87554667 8.45260526,3.63421333 8.2937673,3.42728 C8.13519629,3.22034667 7.93150996,3.05448 7.68270831,2.93048 C7.43390666,2.80594667 7.14426097,2.74408 6.81243645,2.74408 L3.24886005,2.74408 L3.24886005,7.29741333 Z M6.81243645,0.03288 C7.54469281,0.03288 8.19018893,0.15688 8.74972568,0.405413333 C9.30899549,0.653946667 9.7753651,0.998746667 10.1483006,1.44008 C10.5212361,1.88194667 10.803941,2.39928 10.9974831,2.99234667 C11.1912921,3.58568 11.2879296,4.22061333 11.2879296,4.89661333 C11.2879296,5.55874667 11.197699,6.19021333 11.0185725,6.79021333 C10.8389121,7.39048 10.5626141,7.91474667 10.1896786,8.36328 C9.81701001,8.81154667 9.3506404,9.17048 8.79110364,9.43954667 C8.23156689,9.70834667 7.57218913,9.84301333 6.81243645,9.84301333 L3.24886005,9.84301333 L3.24886005,15.8795467 L0.535,15.8795467 L0.535,0.03288 L6.81243645,0.03288 Z" id="Fill-1" fill="#5628B4"></path>
																<polygon id="Fill-37" fill="#5628B4" points="12.755 9.30386667 12.755 0.0329333333 15.46886 0.0329333333 15.46886 10.6366667"></polygon>
																<polygon id="Fill-162" fill="#F7B236" points="15.46886 13.4177067 15.46886 15.8795733 12.755 15.87904 12.755 12.08544"></polygon>
															</g>
														</g>
													</g>
												</g>
											</svg>
									{{.Content}}
									</td>
								</tr>
							</table>
						</td>
					</tr>

				</table>
				<div class="footer">
					<table role="presentation" border="0" cellpadding="0" cellspacing="0">
						<tr>
							<td class="content-block">
								<span class="apple-link">Copyright © 2019 | OpenPitrix | All rights reserved.</span>
							</td>
						</tr>
					</table>
				</div>

			</div>
		</td>
		<td>&nbsp;</td>
	</tr>
</table>
</body>
</html>


`
)
