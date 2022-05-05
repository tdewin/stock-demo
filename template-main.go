package main

const msghtml = `
<!doctype html>
<html lang="en">
	<head>
	<!-- Required meta tags -->
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

	<!-- Bootstrap CSS -->
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
	<title>Product App</title>
	{{ if (ne .Redirect "") }}
	<meta http-equiv="refresh" content="{{.Refresh}}; url={{.Redirect}}">
	{{else if (ne .Refresh 0) }}
	<meta http-equiv="refresh" content="{{.Refresh}}">
	{{end}}
	</head>
	<body class="bg-{{.MessageType}}">
	<div class="container-fluid h-100">
		<div class="row">
				<div class="col-lg-12">&nbsp</div>
		</div>
		<div class="row" >
					<div class="col-lg-1"></div>
                    <div class="col-lg-10">
						<div class="alert alert-{{.MessageType}}" role="alert">
						{{.Pre}} : {{.Message}}
						{{ if (ne .BuySum 0.0) }}

						<table class="table">
					    <tr>
							<th>Product</th>
							<th>Amount</th>
							<th>Price</th>
							<th>Sum</th>
						</tr>
						{{range $i, $t := .BuyTable}}
                        <tr>
							<td>{{$t.Product}}</td>
							<td>{{$t.Bought}}</td>
							<td>{{$t.Price}}</td>
							<td>{{$t.Sum}}</td>
                        </tr>
						{{end}}
						<tr>
							<td colspan=3/>
							<td>{{.BuySum}}</td>
                        </tr>
                      	</table>
						{{end}}
						</div>
						{{ if (ne .LinkBack "") }}
						<div>
							<a href="{{.LinkBack}}" class="btn btn-primary">Continue</a>
						</div>
						{{end}}
					</div>
					<div class="col-lg-1"></div>
		</div>
	</div>
	</body>
</html>
`

const mainhtml = `<!doctype html>
<html lang="en">
	<head>
	<!-- Required meta tags -->
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

	<!-- Bootstrap CSS -->
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
	<title>Product App</title>

	</head>
	<body style="background: #003738;" class="text-light">
		<div class="container-fluid h-100">
			<div class="row align-items-center">
				<div class="col-lg-5"></div>
				<div class="col-lg-2" style="max-width:450px;">
					<div class="m-4">
					<svg
						xmlns:svg="http://www.w3.org/2000/svg"
						xmlns="http://www.w3.org/2000/svg"
						id="svg3858"
						width="100%"
						
						viewBox="0 0 1047 1000"
						version="1.1">
						<path
							style="fill:#007f73;fill-opacity:1;fill-rule:nonzero;stroke:none;stroke-width:15.5041"
							id="path3307"
							d="m 927.80797,482.78598 c 0,0 -218.40538,-406.813364 -218.40538,-406.813364 -23.01388,-42.772503 -75.72283,-58.424394 -117.79501,-34.975162 0,0 -399.73495,222.871966 -399.73495,222.871966 -17.99859,10.29568 -34.78585,28.27552 -42.43203,53.74975 0,0 -107.272348,372.55728 -107.272348,372.55728 -12.850759,44.55175 12.282998,91.27998 56.153349,104.37679 0,0 630.158519,179.2286 630.158519,179.2286 43.83252,13.05882 89.82244,-12.52889 102.69196,-57.1374 0,0 104.24423,-365.30878 104.24423,-365.30878 6.77547,-21.443 2.97137,-47.90158 -7.60834,-68.54968 z" />
						<path
							id="path893"
							d="m 352.70131,404.7392 143.39332,298.35062 201.2132,-359.6397"
							style="fill:none;stroke:#ffffff;stroke-width:100;stroke-linecap:round;stroke-linejoin:bevel;stroke-miterlimit:4;stroke-dasharray:none;stroke-opacity:1" />
					</svg>
					</div>
				</div>
				<div class="col-lg-5"></div>
			</div>
			{{$adm := .Admin}}
			{{ if ( not $adm ) }}
			<form method="post" action="./buy">
			{{else}}
			<form method="post" action="./set">
			{{end}}
			<div class="row" >
					<div class="col-lg-2"></div>
					
                    <div class="col-lg-8">
                      <table class="table">
					    <tr>

							
							{{ if ( $adm ) }}
							<th></th>
							<th>Product</th>
							<th>Availability</th>
							<th></th>
							<th>Price</th>
							{{else}}
							<th>Product</th>
							<th>Availability</th>
							<th>Price</th>
							<th></th>
							{{ end }}
						</tr>

						{{range .Stocks}}
                        <tr>
						  
						  <td>
						  {{ if ( $adm ) }}
						  <select name="setkeep-{{.BuyID}}" class="form-control my-0">
  								<option value="keep" selected>Keep</option>
							    <option value="delete">Delete</option>
						  </select>
						  </td>
						  <td>
						  <input class="form-control my-0" type="text"  value="{{.Product}}" name="setproduct-{{.BuyID}}"/>
						  {{ else }} 
						  	{{.Product}}
						  {{ end}}
						  </td>

						  <td>
						  {{ if ( $adm ) }}
						   <input class="form-control my-0" type="text"  value="{{.Stock}}" name="setstock-{{.BuyID}}"/>
						  </td>
						  <td>
						   <input class="form-control my-0" type="text"  value="{{.Unit}}" name="setunit-{{.BuyID}}"/>
						  {{else if (gt .Stock 0.0) }}  
							{{.StockMessage}}
						  {{else}}
							Out of stock!
						  {{end}}
						  </td>

						  <td>
						  {{ if ( $adm ) }}
						  <input class="form-control my-0" type="text"  value="{{.Price}}" name="setprice-{{.BuyID}}"/>
						  {{ else if (gt .Stock 0.0) }} 
						    {{.Price}}
						  {{end}}
						  </td>

						  {{ if ( not $adm ) }}
						  <td  class="text-right py-2">
						  {{ if (gt .Stock 0.0) }} 
						    <input class="form-control my-0" type="text" max="{{.Stock}}" value="0" name="qty-{{.BuyID}}"/>
						  {{end}}
						  </td>
						  {{end}}

                        </tr>
						{{end}}
						<tr>
						
						{{ if ( not $adm ) }}
						<td colspan=3></td>
						<td  class="text-right"><button class="btn btn-primary btn-sm">Buy</button></td>
						{{else}}
						<td colspan=4></td>
						<td  class="text-right"><button class="btn btn-primary btn-sm">Set</button></td>
						{{end}}
						
						</tr>
                      </table>
					  
                    </div>
					<div class="col-lg-2"></div>
            </div>
			</form>
		</div>
				
		<!-- Optional JavaScript -->
		<!-- jQuery first, then Popper.js, then Bootstrap JS -->
		<script src="https://code.jquery.com/jquery-3.5.1.min.js" integrity="sha384-ZvpUoO/+PpLXR1lu4jmpXWu80pZlYUAfxl5NsBMWOEPSjUn/6Z/hRTt8+pR6L4N2" crossorigin="anonymous"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js" integrity="sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q" crossorigin="anonymous"></script>
		<script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js" integrity="sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl" crossorigin="anonymous"></script>
	</body>
</html>
`
